# Chapter 3: Creating a Container

## 1. Creating a Container with run Command

* The goal is to realize a simple run command, similar to ```docker run -ti [commnad]```.

* Code structure:

    ```console
    mydocker/
    ├── container
    │   ├── container_process.go
    │   └── init.go
    ├── main_command.go
    ├── main.go
    ├── mydocker
    └── run.go
    ```

* MountFlag
    * ```MS_NOEXEC```: running other programs in this fs is not allowed
    * ```MS_NOSUID```: cannot set user-ID or group-ID while running a process in this fs
    * ```MS_NODEV```: all mount systems must set this since Linux kernel 2.4

* ```syscall.Exec``` calls kernel's ```int execve(const char *filename, char *const argv[], char *const envp[]);```, which executes a binary with the name ```filename``` and overwrites the image, data, heap/stack of the current process, including the PID. This essentially substitutes the init process. Inside the container, the first process will be what we specify, instead of init.

* Program flow:
![mydocker](../resources/ch3_1.jpg)

* Test (with root access):

    ```console
        $ go build -o mydocker .

        $ ./mydocker run -ti /bin/sh
        {"level":"info","msg":"init come on","source":"3.1/main_command.go:51","time":"2018-07-11T17:41:06+08:00"}
        {"level":"info","msg":"command /bin/sh","source":"3.1/main_command.go:53","time":"2018-07-11T17:41:06+08:00"}
        {"level":"info","msg":"command /bin/sh","source":"container/init.go:17","time":"2018-07-11T17:41:06+08:00"}
        # ps -ef
        UID        PID  PPID  C STIME TTY          TIME CMD
        root         1     0  0 17:41 pts/18   00:00:00 /bin/sh
        root         6     1  0 17:41 pts/18   00:00:00 ps -ef
        # exit

        $ ./mydocker run -ti /bin/ls
        {"level":"info","msg":"init come on","source":"3.1/main_command.go:51","time":"2018-07-11T17:41:21+08:00"}
        {"level":"info","msg":"command /bin/ls","source":"3.1/main_command.go:53","time":"2018-07-11T17:41:21+08:00"}
        {"level":"info","msg":"command /bin/ls","source":"container/init.go:17","time":"2018-07-11T17:41:21+08:00"}
        container  main_command.go  main.go  mydocker  run.go
    ```