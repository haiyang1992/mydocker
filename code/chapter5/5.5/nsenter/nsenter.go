package nsenter

/*
#define _GNU_SOURCE
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>


// the __attribute__((constructor)) here means that once the package is called, the function will run automatically
// similar to a constructor running at the beginning of the program
__attribute__((constructor)) void enter_namespace(void) {
    char *mydocker_pid;
    // get PID from environmental variable
    mydocker_pid = getenv("mydocker_pid");
    if (mydocker_pid){
        fprintf(stdout, "nsenter: got mydocker_pid = %s\n", mydocker_pid);
    }
    else{
        fprintf(stdout, "nsenter: missing mydocker_pid env, skipping nsenter\n");
        return;
    }
    char *mydocker_cmd;
    mydocker_cmd = getenv("mydocker_cmd");
    if (mydocker_cmd){
        fprintf(stdout, "nsenter: got mydocker_cmd = %s\n", mydocker_cmd);
    }
    else{
        fprintf(stdout, "nsenter: missing mydocker_cmd env, skipping nsenter\n");
        return;
    }

    int i;
    char nspath[1024];
    // the five Namespaces we need to enter
    char *namespaces[] = {"ipc", "uts", "net", "pid", "mnt"};

    for (i=0;i<5;i++){
        // piece together the corresponding path
        sprintf(nspath, "/proc/%s/ns/%s", mydocker_pid, namespaces[i]);
        int fd = open(nspath, O_RDONLY);
        // now we call the setns syscall
        if (setns(fd, 0) == -1){
            fprintf(stderr, "nsenter: setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
        }
        else{
            fprintf(stdout, "nsenter: setns on %s namespace success\n", namespaces[i]);
        }
        close(fd);
    }
    // run the designaetd command within namespace
    int res = system(mydocker_cmd);
    exit(0);
    return;
}
*/
import "C"
