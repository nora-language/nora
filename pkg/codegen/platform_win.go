package codegen

import "fmt"

type WindowsTarget struct{}

func (t *WindowsTarget) Name() string { return "windows" }

func (t *WindowsTarget) Bootstrap() string {
	return `
    nr_argc = argc;
    nr_argv = argv;
    setvbuf(stdout, NULL, _IONBF, 0);
    nr_mem_init();
    scheduler_init();
    nr_init_globals();
`
}

func (t *WindowsTarget) GetMainWrapper(mainFunc string) string {
	return fmt.Sprintf(`
void nr_cleanup_globals(void);

static void __nr_main_wrapper(void* p) {
    %s(NULL);
    nr_flush_temps();
}

int main(int argc, char** argv) {
    %s
    scheduler_spawn(__nr_main_wrapper, NULL, "main", "main", 0);
    scheduler_run_loop();
    g_running = false;
    scheduler_cleanup();
#ifdef NR_DEBUG_FIBER
    nr_fiber_report();
#endif
    nr_cleanup_globals();
    nr_mem_report();
    return 0;
}`, mainFunc, t.Bootstrap())
}
