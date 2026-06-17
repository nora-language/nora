#include "runtime.h"

// --- PREMIUM NET RUNTIME ---
#ifdef __wasm__
// WASM / WASI has no standard socket stack by default, so we stub them out to prevent build failures.
#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

int32_t nr_net_dial(char* addr, int32_t port) { return -3; }
int32_t nr_net_listen(int32_t port) { return -1; }
int32_t nr_net_accept(int32_t listener_fd) { return -1; }
int32_t nr_net_send(int32_t socket_fd, unsigned char* data, int32_t len) { return -1; }
unsigned char* nr_net_recv(int32_t socket_fd, int32_t max_len) { return NULL; }
void nr_net_close(int32_t socket_fd) {}
int32_t nr_net_udp_socket(void) { return -1; }
int32_t nr_net_udp_bind(int32_t socket_fd, int32_t port) { return -1; }
int32_t nr_net_udp_sendto(int32_t socket_fd, unsigned char* data, int32_t len, char* addr, int32_t port) { return -1; }
unsigned char* nr_net_udp_recvfrom(int32_t socket_fd, int32_t max_len, unsigned char* out_addr, int32_t* out_port) { return NULL; }

#else

#ifdef _WIN32
    #include <winsock2.h>
    #include <ws2tcpip.h>
    #include <stddef.h>
#else
    #include <sys/socket.h>
    #include <netinet/in.h>
    #include <arpa/inet.h>
    #include <unistd.h>
    #include <netdb.h>
    #include <fcntl.h>
    #include <errno.h>
#ifdef __linux__
    #include <sys/epoll.h>
#endif
#endif

typedef struct net_waiter {
    int32_t fd;
    fiber_info_t* fiber;
    bool is_write;
#ifdef _WIN32
    WSAOVERLAPPED overlapped;
    WSABUF wsa_buf;
    HANDLE wsa_event;
    HANDLE wait_handle;
#endif
    struct net_waiter* next;
} net_waiter_t;

static net_waiter_t* g_net_waiters_head = NULL;
static NR_MUTEX_T g_net_waiters_lock;
static bool g_net_poller_running = false;

#ifdef _WIN32
static HANDLE g_iocp = INVALID_HANDLE_VALUE;
#elif defined(__linux__)
static int g_epoll_fd = -1;
#endif

#ifdef _WIN32
static void nr_net_waiter_cleanup(net_waiter_t* waiter) {
    if (waiter->wsa_event != NULL && waiter->wsa_event != INVALID_HANDLE_VALUE) {
        UnregisterWait(waiter->wait_handle);
        WSAEventSelect((SOCKET)waiter->fd, NULL, 0);
        WSACloseEvent(waiter->wsa_event);
        waiter->wsa_event = NULL;
    }
}
#else
static void nr_net_waiter_cleanup(net_waiter_t* waiter) {}
#endif

static bool nr_net_waiter_remove(net_waiter_t* waiter) {
    NR_MUTEX_LOCK(&g_net_waiters_lock);
    net_waiter_t* curr = g_net_waiters_head;
    net_waiter_t* prev = NULL;
    bool found = false;
    while (curr) {
        if (curr == waiter) {
            if (prev) {
                prev->next = curr->next;
            } else {
                g_net_waiters_head = curr->next;
            }
            found = true;
            break;
        }
        prev = curr;
        curr = curr->next;
    }
    NR_MUTEX_UNLOCK(&g_net_waiters_lock);
    if (found) {
        nr_net_waiter_cleanup(waiter);
    }
    return found;
}

static void nr_net_register_socket(int32_t fd) {
#ifdef _WIN32
    if (g_iocp != INVALID_HANDLE_VALUE) {
        CreateIoCompletionPort((HANDLE)(intptr_t)fd, g_iocp, 0, 0);
    }
#endif
}

#ifdef _WIN32
static DWORD WINAPI nr_net_poller_thread(LPVOID arg) {
    while (g_net_poller_running) {
        DWORD bytes_transferred = 0;
        ULONG_PTR completion_key = 0;
        LPOVERLAPPED overlapped = NULL;
        BOOL ok = GetQueuedCompletionStatus(g_iocp, &bytes_transferred, &completion_key, &overlapped, INFINITE);
        if (!ok && overlapped == NULL) {
            break;
        }
        if (completion_key == 9999) {
            break;
        }
        if (overlapped) {
            net_waiter_t* waiter = (net_waiter_t*)((char*)overlapped - offsetof(net_waiter_t, overlapped));
            if (nr_net_waiter_remove(waiter)) {
                NR_ATOMIC_SUB(&g_net_waiters_count, 1);
                resume(waiter->fiber);
                free(waiter);
            }
        }
    }
    return 0;
}
#elif defined(__linux__)
static void* nr_net_poller_thread(void* arg) {
    #define MAX_EVENTS 64
    struct epoll_event events[MAX_EVENTS];
    while (g_net_poller_running) {
        int nfds = epoll_wait(g_epoll_fd, events, MAX_EVENTS, -1);
        if (nfds < 0) {
            if (errno == EINTR) continue;
            break;
        }
        for (int i = 0; i < nfds; i++) {
            net_waiter_t* waiter = (net_waiter_t*)events[i].data.ptr;
            if (waiter) {
                if (nr_net_waiter_remove(waiter)) {
                    epoll_ctl(g_epoll_fd, EPOLL_CTL_DEL, waiter->fd, NULL);
                    NR_ATOMIC_SUB(&g_net_waiters_count, 1);
                    resume(waiter->fiber);
                    free(waiter);
                }
            }
        }
    }
    return NULL;
}
#else
static void* nr_net_poller_thread(void* arg) {
    while (g_net_poller_running) {
        NR_MUTEX_LOCK(&g_net_waiters_lock);
        net_waiter_t* curr = g_net_waiters_head;
        net_waiter_t* prev = NULL;
        while (curr) {
            bool ready = false;
            fd_set fds;
            FD_ZERO(&fds);
            FD_SET(curr->fd, &fds);
            struct timeval tv = {0, 0};
            int sel;
            if (curr->is_write) {
                sel = select(curr->fd + 1, NULL, &fds, NULL, &tv);
            } else {
                sel = select(curr->fd + 1, &fds, NULL, NULL, &tv);
            }
            if (sel > 0 && FD_ISSET(curr->fd, &fds)) {
                ready = true;
            } else if (sel < 0) {
                ready = true;
            }

            if (ready) {
                NR_ATOMIC_SUB(&g_net_waiters_count, 1);
                resume(curr->fiber);
                net_waiter_t* next = curr->next;
                if (prev) {
                    prev->next = next;
                } else {
                    g_net_waiters_head = next;
                }
                free(curr);
                curr = next;
            } else {
                prev = curr;
                curr = curr->next;
            }
        }
        bool has_waiters = (g_net_waiters_head != NULL);
        NR_MUTEX_UNLOCK(&g_net_waiters_lock);

        if (has_waiters) {
            usleep(100);
        } else {
            usleep(10000);
        }
    }
    return NULL;
}
#endif

static void nr_set_nonblocking(int32_t fd) {
#ifdef _WIN32
    u_long mode = 1;
    ioctlsocket((SOCKET)fd, FIONBIO, &mode);
#else
    int flags = fcntl(fd, F_GETFL, 0);
    if (flags != -1) {
        fcntl(fd, F_SETFL, flags | O_NONBLOCK);
    }
#endif
}

static void nr_net_init(void) {
#ifdef _WIN32
    static bool initialized = false;
    if (!initialized) {
        WSADATA wsa;
        if (WSAStartup(MAKEWORD(2, 2), &wsa) == 0) {
            initialized = true;
        }
    }
#endif
    static bool poller_initialized = false;
    if (!poller_initialized) {
        NR_MUTEX_INIT(&g_net_waiters_lock);
        g_net_poller_running = true;
#ifdef _WIN32
        g_iocp = CreateIoCompletionPort(INVALID_HANDLE_VALUE, NULL, 0, 0);
        CreateThread(NULL, 0, nr_net_poller_thread, NULL, 0, NULL);
#elif defined(__linux__)
        g_epoll_fd = epoll_create1(0);
        pthread_t tid;
        pthread_create(&tid, NULL, nr_net_poller_thread, NULL);
#else
#ifdef _WIN32
        CreateThread(NULL, 0, nr_net_poller_thread, NULL, 0, NULL);
#else
        pthread_t tid;
        pthread_create(&tid, NULL, nr_net_poller_thread, NULL);
#endif
#endif
        poller_initialized = true;
    }
}

int32_t nr_net_dial(char* addr, int32_t port) {
    nr_net_init();
#ifdef _WIN32
    SOCKET s = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    if (s == INVALID_SOCKET) return -1;
#else
    int s = socket(AF_INET, SOCK_STREAM, 0);
    if (s < 0) return -1;
#endif

    struct hostent* he = gethostbyname(addr);
    if (he == NULL) {
#ifdef _WIN32
        closesocket(s);
#else
        close(s);
#endif
        return -2;
    }

    struct sockaddr_in server;
    memset(&server, 0, sizeof(server));
    server.sin_family = AF_INET;
    server.sin_port = htons(port);
    memcpy(&server.sin_addr, he->h_addr_list[0], he->h_length);

    if (connect(s, (struct sockaddr*)&server, sizeof(server)) < 0) {
#ifdef _WIN32
        closesocket(s);
#else
        close(s);
#endif
        return -3;
    }

    nr_set_nonblocking((int32_t)s);
    nr_net_register_socket((int32_t)s);
    return (int32_t)s;
}

int32_t nr_net_listen(int32_t port) {
    nr_net_init();
#ifdef _WIN32
    SOCKET s = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    if (s == INVALID_SOCKET) return -1;
#else
    int s = socket(AF_INET, SOCK_STREAM, 0);
    if (s < 0) return -1;
#endif

    int opt = 1;
#ifdef _WIN32
    setsockopt(s, SOL_SOCKET, SO_REUSEADDR, (char*)&opt, sizeof(opt));
#else
    setsockopt(s, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
#endif

    struct sockaddr_in server;
    memset(&server, 0, sizeof(server));
    server.sin_family = AF_INET;
    server.sin_addr.s_addr = INADDR_ANY;
    server.sin_port = htons(port);

    if (bind(s, (struct sockaddr*)&server, sizeof(server)) < 0) {
#ifdef _WIN32
        closesocket(s);
#else
        close(s);
#endif
        return -2;
    }

    if (listen(s, 10) < 0) {
#ifdef _WIN32
        closesocket(s);
#else
        close(s);
#endif
        return -3;
    }

    nr_set_nonblocking((int32_t)s);
    nr_net_register_socket((int32_t)s);
    return (int32_t)s;
}

#ifdef _WIN32
static VOID CALLBACK nr_net_accept_callback(PVOID lpParameter, BOOLEAN TimerOrWaitFired) {
    net_waiter_t* waiter = (net_waiter_t*)lpParameter;
    if (waiter) {
        PostQueuedCompletionStatus(g_iocp, 0, 0, &waiter->overlapped);
    }
}
#endif

int32_t nr_net_accept(int32_t listener_fd) {
    nr_net_init();
    nr_set_nonblocking(listener_fd);
    struct sockaddr_in client;
    int client_len = sizeof(client);
    while (1) {
#ifdef _WIN32
        SOCKET client_fd = accept((SOCKET)listener_fd, (struct sockaddr*)&client, &client_len);
        if (client_fd != INVALID_SOCKET) {
            nr_set_nonblocking((int32_t)client_fd);
            nr_net_register_socket((int32_t)client_fd);
            return (int32_t)client_fd;
        }
        int err = WSAGetLastError();
        if (err != WSAEWOULDBLOCK) return -1;
#else
        int client_fd = accept(listener_fd, (struct sockaddr*)&client, (socklen_t*)&client_len);
        if (client_fd >= 0) {
            nr_set_nonblocking(client_fd);
            nr_net_register_socket(client_fd);
            return client_fd;
        }
        if (errno != EAGAIN && errno != EWOULDBLOCK) return -1;
#endif

        fiber_info_t* self = (fiber_info_t*)GetFiberData();
        if (!self) {
#ifdef _WIN32
            Sleep(1);
#else
            usleep(1000);
#endif
            continue;
        }

        net_waiter_t* waiter = (net_waiter_t*)malloc(sizeof(net_waiter_t));
        waiter->fd = listener_fd;
        waiter->fiber = self;
        waiter->is_write = false;
        waiter->next = NULL;
#ifdef _WIN32
        waiter->wsa_event = NULL;
        waiter->wait_handle = NULL;
#endif

        NR_MUTEX_LOCK(&g_net_waiters_lock);
        waiter->next = g_net_waiters_head;
        g_net_waiters_head = waiter;
        NR_MUTEX_UNLOCK(&g_net_waiters_lock);

        NR_ATOMIC_ADD(&g_net_waiters_count, 1);

#ifdef _WIN32
        memset(&waiter->overlapped, 0, sizeof(WSAOVERLAPPED));
        waiter->wsa_buf.buf = NULL;
        waiter->wsa_buf.len = 0;
        waiter->wsa_event = WSACreateEvent();
        WSAEventSelect((SOCKET)listener_fd, waiter->wsa_event, FD_ACCEPT | FD_CLOSE);
        RegisterWaitForSingleObject(&waiter->wait_handle, waiter->wsa_event, nr_net_accept_callback, waiter, INFINITE, WT_EXECUTEONLYONCE);
#elif defined(__linux__)
        struct epoll_event ev;
        memset(&ev, 0, sizeof(ev));
        ev.events = EPOLLIN | EPOLLONESHOT | EPOLLET;
        ev.data.ptr = waiter;
        if (epoll_ctl(g_epoll_fd, EPOLL_CTL_ADD, listener_fd, &ev) < 0) {
            if (errno == EEXIST) {
                epoll_ctl(g_epoll_fd, EPOLL_CTL_MOD, listener_fd, &ev);
            }
        }
#endif

        park();
    }
}

int32_t nr_net_send(int32_t socket_fd, unsigned char* data, int32_t len) {
    nr_net_init();
    nr_set_nonblocking(socket_fd);
    while (1) {
        int n = send(socket_fd, (char*)data, len, 0);
        if (n >= 0) return n;
#ifdef _WIN32
        int err = WSAGetLastError();
        if (err != WSAEWOULDBLOCK) return -1;
#else
        if (errno != EAGAIN && errno != EWOULDBLOCK) return -1;
#endif

        fiber_info_t* self = (fiber_info_t*)GetFiberData();
        if (!self) {
#ifdef _WIN32
            Sleep(1);
#else
            usleep(1000);
#endif
            continue;
        }

        net_waiter_t* waiter = (net_waiter_t*)malloc(sizeof(net_waiter_t));
        waiter->fd = socket_fd;
        waiter->fiber = self;
        waiter->is_write = true;
        waiter->next = NULL;
#ifdef _WIN32
        waiter->wsa_event = NULL;
        waiter->wait_handle = NULL;
#endif

        NR_MUTEX_LOCK(&g_net_waiters_lock);
        waiter->next = g_net_waiters_head;
        g_net_waiters_head = waiter;
        NR_MUTEX_UNLOCK(&g_net_waiters_lock);

        NR_ATOMIC_ADD(&g_net_waiters_count, 1);

#ifdef _WIN32
        memset(&waiter->overlapped, 0, sizeof(WSAOVERLAPPED));
        waiter->wsa_buf.buf = NULL;
        waiter->wsa_buf.len = 0;
        DWORD bytes = 0;
        WSASend((SOCKET)socket_fd, &waiter->wsa_buf, 1, &bytes, 0, &waiter->overlapped, NULL);
#elif defined(__linux__)
        struct epoll_event ev;
        memset(&ev, 0, sizeof(ev));
        ev.events = EPOLLOUT | EPOLLONESHOT | EPOLLET;
        ev.data.ptr = waiter;
        if (epoll_ctl(g_epoll_fd, EPOLL_CTL_ADD, socket_fd, &ev) < 0) {
            if (errno == EEXIST) {
                epoll_ctl(g_epoll_fd, EPOLL_CTL_MOD, socket_fd, &ev);
            }
        }
#endif

        park();
    }
}

unsigned char* nr_net_recv(int32_t socket_fd, int32_t max_len) {
    if (max_len <= 0) return NULL;
    nr_net_init();
    nr_set_nonblocking(socket_fd);
    unsigned char* buf = nr_malloc_debug(max_len, "std/net/net.nr", 0);
    while (1) {
        int n = recv(socket_fd, (char*)buf, max_len, 0);
        if (n > 0) {
            nr_header_t* h = (nr_header_t*)((char*)buf - NR_HEADER_SIZE);
            h->elem_size = 1;
            h->count = n;
            return buf;
        }
        if (n == 0) {
            nr_free(buf);
            return NULL;
        }
#ifdef _WIN32
        int err = WSAGetLastError();
        if (err != WSAEWOULDBLOCK) {
            nr_free(buf);
            return NULL;
        }
#else
        if (errno != EAGAIN && errno != EWOULDBLOCK) {
            nr_free(buf);
            return NULL;
        }
#endif

        fiber_info_t* self = (fiber_info_t*)GetFiberData();
        if (!self) {
#ifdef _WIN32
            Sleep(1);
#else
            usleep(1000);
#endif
            continue;
        }

        net_waiter_t* waiter = (net_waiter_t*)malloc(sizeof(net_waiter_t));
        waiter->fd = socket_fd;
        waiter->fiber = self;
        waiter->is_write = false;
        waiter->next = NULL;
#ifdef _WIN32
        waiter->wsa_event = NULL;
        waiter->wait_handle = NULL;
#endif

        NR_MUTEX_LOCK(&g_net_waiters_lock);
        waiter->next = g_net_waiters_head;
        g_net_waiters_head = waiter;
        NR_MUTEX_UNLOCK(&g_net_waiters_lock);

        NR_ATOMIC_ADD(&g_net_waiters_count, 1);

#ifdef _WIN32
        memset(&waiter->overlapped, 0, sizeof(WSAOVERLAPPED));
        waiter->wsa_buf.buf = NULL;
        waiter->wsa_buf.len = 0;
        DWORD bytes = 0;
        DWORD flags = 0;
        WSARecv((SOCKET)socket_fd, &waiter->wsa_buf, 1, &bytes, &flags, &waiter->overlapped, NULL);
#elif defined(__linux__)
        struct epoll_event ev;
        memset(&ev, 0, sizeof(ev));
        ev.events = EPOLLIN | EPOLLONESHOT | EPOLLET;
        ev.data.ptr = waiter;
        if (epoll_ctl(g_epoll_fd, EPOLL_CTL_ADD, socket_fd, &ev) < 0) {
            if (errno == EEXIST) {
                epoll_ctl(g_epoll_fd, EPOLL_CTL_MOD, socket_fd, &ev);
            }
        }
#endif

        park();
    }
}

void nr_net_close(int32_t socket_fd) {
    NR_MUTEX_LOCK(&g_net_waiters_lock);
    net_waiter_t* curr = g_net_waiters_head;
    net_waiter_t* prev = NULL;
    while (curr) {
        if (curr->fd == socket_fd) {
            net_waiter_t* to_free = curr;
            curr = curr->next;
            if (prev) {
                prev->next = curr;
            } else {
                g_net_waiters_head = curr;
            }
#if defined(__linux__)
            epoll_ctl(g_epoll_fd, EPOLL_CTL_DEL, to_free->fd, NULL);
#endif
            nr_net_waiter_cleanup(to_free);
            NR_ATOMIC_SUB(&g_net_waiters_count, 1);
            resume(to_free->fiber);
            free(to_free);
        } else {
            prev = curr;
            curr = curr->next;
        }
    }
    NR_MUTEX_UNLOCK(&g_net_waiters_lock);

#ifdef _WIN32
    closesocket((SOCKET)socket_fd);
#else
    close(socket_fd);
#endif
}

int32_t nr_net_udp_socket(void) {
    nr_net_init();
#ifdef _WIN32
    SOCKET s = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
    if (s == INVALID_SOCKET) return -1;
#else
    int s = socket(AF_INET, SOCK_DGRAM, 0);
    if (s < 0) return -1;
#endif
    nr_set_nonblocking((int32_t)s);
    nr_net_register_socket((int32_t)s);
    return (int32_t)s;
}

int32_t nr_net_udp_bind(int32_t socket_fd, int32_t port) {
    struct sockaddr_in server;
    memset(&server, 0, sizeof(server));
    server.sin_family = AF_INET;
    server.sin_addr.s_addr = INADDR_ANY;
    server.sin_port = htons(port);
#ifdef _WIN32
    if (bind((SOCKET)socket_fd, (struct sockaddr*)&server, sizeof(server)) < 0) {
        return -1;
    }
#else
    if (bind(socket_fd, (struct sockaddr*)&server, sizeof(server)) < 0) {
        return -1;
    }
#endif
    return 0;
}

int32_t nr_net_udp_sendto(int32_t socket_fd, unsigned char* data, int32_t len, char* addr, int32_t port) {
    struct hostent* he = gethostbyname(addr);
    if (he == NULL) return -2;
    struct sockaddr_in dest;
    memset(&dest, 0, sizeof(dest));
    dest.sin_family = AF_INET;
    dest.sin_port = htons(port);
    memcpy(&dest.sin_addr, he->h_addr_list[0], he->h_length);
#ifdef _WIN32
    int n = sendto((SOCKET)socket_fd, (char*)data, len, 0, (struct sockaddr*)&dest, sizeof(dest));
#else
    int n = sendto(socket_fd, (char*)data, len, 0, (struct sockaddr*)&dest, sizeof(dest));
#endif
    return n;
}

unsigned char* nr_net_udp_recvfrom(int32_t socket_fd, int32_t max_len, unsigned char* out_addr, int32_t* out_port) {
    if (max_len <= 0) return NULL;
    nr_net_init();
    nr_set_nonblocking(socket_fd);
    unsigned char* buf = nr_malloc_debug(max_len, "std/net/net.nr", 0);
    struct sockaddr_in client;
    int client_len = sizeof(client);
    while (1) {
#ifdef _WIN32
        int n = recvfrom((SOCKET)socket_fd, (char*)buf, max_len, 0, (struct sockaddr*)&client, &client_len);
#else
        int n = recvfrom(socket_fd, (char*)buf, max_len, 0, (struct sockaddr*)&client, (socklen_t*)&client_len);
#endif
        if (n > 0) {
            nr_header_t* h = (nr_header_t*)((char*)buf - NR_HEADER_SIZE);
            h->elem_size = 1;
            h->count = n;
            char* ip = inet_ntoa(client.sin_addr);
            if (ip && out_addr) {
                strcpy((char*)out_addr, ip);
            }
            if (out_port) {
                *out_port = ntohs(client.sin_port);
            }
            return buf;
        }
        if (n == 0) {
            nr_free(buf);
            return NULL;
        }
#ifdef _WIN32
        int err = WSAGetLastError();
        if (err != WSAEWOULDBLOCK) {
            nr_free(buf);
            return NULL;
        }
#else
        if (errno != EAGAIN && errno != EWOULDBLOCK) {
            nr_free(buf);
            return NULL;
        }
#endif

        fiber_info_t* self = (fiber_info_t*)GetFiberData();
        if (!self) {
#ifdef _WIN32
            Sleep(1);
#else
            usleep(1000);
#endif
            continue;
        }

        net_waiter_t* waiter = (net_waiter_t*)malloc(sizeof(net_waiter_t));
        waiter->fd = socket_fd;
        waiter->fiber = self;
        waiter->is_write = false;
        waiter->next = NULL;
#ifdef _WIN32
        waiter->wsa_event = NULL;
        waiter->wait_handle = NULL;
#endif

        NR_MUTEX_LOCK(&g_net_waiters_lock);
        waiter->next = g_net_waiters_head;
        g_net_waiters_head = waiter;
        NR_MUTEX_UNLOCK(&g_net_waiters_lock);

        NR_ATOMIC_ADD(&g_net_waiters_count, 1);

#ifdef _WIN32
        memset(&waiter->overlapped, 0, sizeof(WSAOVERLAPPED));
        waiter->wsa_buf.buf = NULL;
        waiter->wsa_buf.len = 0;
        waiter->wsa_event = WSACreateEvent();
        WSAEventSelect((SOCKET)socket_fd, waiter->wsa_event, FD_READ | FD_CLOSE);
        RegisterWaitForSingleObject(&waiter->wait_handle, waiter->wsa_event, nr_net_accept_callback, waiter, INFINITE, WT_EXECUTEONLYONCE);
#elif defined(__linux__)
        struct epoll_event ev;
        memset(&ev, 0, sizeof(ev));
        ev.events = EPOLLIN | EPOLLONESHOT | EPOLLET;
        ev.data.ptr = waiter;
        if (epoll_ctl(g_epoll_fd, EPOLL_CTL_ADD, socket_fd, &ev) < 0) {
            if (errno == EEXIST) {
                epoll_ctl(g_epoll_fd, EPOLL_CTL_MOD, socket_fd, &ev);
            }
        }
#endif

        park();
    }
}
#endif
