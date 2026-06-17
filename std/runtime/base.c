
int nr_argc = 0;
char** nr_argv = NULL;

void* nr_get_stdout() { return stdout; }
void* nr_get_stderr() { return stderr; }
void* nr_get_stdin() { return stdin; }
int nr_atoi(char* s) { return atoi(s); }
double nr_atof(char* s) { return atof(s); }
