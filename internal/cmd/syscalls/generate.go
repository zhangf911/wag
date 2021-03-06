// Copyright (c) 2018 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"
)

const (
	ptr1 = (1 << 0)
	ptr2 = (1 << 1)
	ptr3 = (1 << 2)
	ptr4 = (1 << 3)
	ptr5 = (1 << 4)
	ptr6 = (1 << 5)
)

// wag's internal calling convention
var regs = []string{"CX", "BX", "BP", "SI", "DI", "R8"}

type call struct {
	name    string
	number  int
	ptrMask int
}

func (sc call) titleName() string {
	return strings.Replace(strings.Title(strings.Replace(sc.name, "_", " ", -1)), " ", "", -1)
}

func main() {
	decl, err := os.Create("syscall.go")
	if err != nil {
		log.Panic(err)
	}
	defer decl.Close()

	impl, err := os.Create("syscall_asm64.s")
	if err != nil {
		log.Panic(err)
	}
	defer impl.Close()

	fmt.Fprintf(decl, "// Generated by internal/cmd/syscalls/generate.go\n\n")
	fmt.Fprintf(decl, "package main\n\n")

	fmt.Fprintf(impl, "// Generated by internal/cmd/syscalls/generate.go\n\n")
	fmt.Fprintf(impl, "#include \"textflag.h\"\n")

	for _, sc := range syscalls {
		fmt.Fprintf(decl, "func import%s() uint64\n", sc.titleName())

		fmt.Fprintf(impl, "\n// func import%s() uint64\n", sc.titleName())
		fmt.Fprintf(impl, "TEXT ·import%s(SB),$0-8\n", sc.titleName())
		fmt.Fprintf(impl, "\tLEAQ\tsys%s<>(SB), AX\n", sc.titleName())
		fmt.Fprintf(impl, "\tMOVQ\tAX, ret+0(FP)\n")
		fmt.Fprintf(impl, "\tRET\n\n")

		fmt.Fprintf(impl, "TEXT sys%s<>(SB),NOSPLIT,$0\n", sc.titleName())

		for i, reg := range regs {
			if (sc.ptrMask & (1 << uint(i))) != 0 {
				fmt.Fprintf(impl, "\tANDL\t%s, %s\n", reg, reg) // zero-extend and test
				fmt.Fprintf(impl, "\tJZ\tnull%d\n", i)
				fmt.Fprintf(impl, "\tADDQ\tR14, %s\n", reg)
				fmt.Fprintf(impl, "null%d:", i)
			}
		}

		fmt.Fprintf(impl, "\tMOVL\t$%d, AX\n", sc.number)
		fmt.Fprintf(impl, "\tJMP\t·callSys(SB)\n")
	}

	fmt.Fprintf(decl, "\nfunc init() {\n")

	for _, sc := range syscalls {
		fmt.Fprintf(decl, "\timportFuncs[\"%s\"] = import%s()\n", sc.name, sc.titleName())
	}

	fmt.Fprintf(decl, "}\n") // init()
}

var syscalls = []call{
	{"read", syscall.SYS_READ, ptr2},
	{"write", syscall.SYS_WRITE, ptr2},
	{"open", syscall.SYS_OPEN, ptr1},
	{"close", syscall.SYS_CLOSE, 0},
	{"lseek", syscall.SYS_LSEEK, 0},
	{"pread", syscall.SYS_PREAD64, ptr2},
	{"pwrite", syscall.SYS_PWRITE64, ptr2},
	{"access", syscall.SYS_ACCESS, ptr1},
	{"pipe", syscall.SYS_PIPE, ptr1},
	{"dup", syscall.SYS_DUP, 0},
	{"dup2", syscall.SYS_DUP2, 0},
	{"getpid", syscall.SYS_GETPID, 0},
	{"sendfile", syscall.SYS_SENDFILE, ptr3},
	{"shutdown", syscall.SYS_SHUTDOWN, 0},
	{"socketpair", syscall.SYS_SOCKETPAIR, ptr4},
	{"flock", syscall.SYS_FLOCK, 0},
	{"fsync", syscall.SYS_FSYNC, 0},
	{"fdatasync", syscall.SYS_FDATASYNC, 0},
	{"truncate", syscall.SYS_TRUNCATE, ptr1},
	{"ftruncate", syscall.SYS_FTRUNCATE, 0},
	{"getcwd", syscall.SYS_GETCWD, ptr1},
	{"chdir", syscall.SYS_CHDIR, ptr1},
	{"fchdir", syscall.SYS_FCHDIR, 0},
	{"rename", syscall.SYS_RENAME, ptr1 | ptr2},
	{"mkdir", syscall.SYS_MKDIR, ptr1},
	{"rmdir", syscall.SYS_RMDIR, ptr1},
	{"creat", syscall.SYS_CREAT, ptr1},
	{"link", syscall.SYS_LINK, ptr1 | ptr2},
	{"unlink", syscall.SYS_UNLINK, ptr1},
	{"symlink", syscall.SYS_SYMLINK, ptr1 | ptr2},
	{"readlink", syscall.SYS_READLINK, ptr1 | ptr2},
	{"chmod", syscall.SYS_CHMOD, ptr1},
	{"fchmod", syscall.SYS_FCHMOD, 0},
	{"chown", syscall.SYS_CHOWN, ptr1},
	{"fchown", syscall.SYS_FCHOWN, 0},
	{"lchown", syscall.SYS_LCHOWN, ptr1},
	{"umask", syscall.SYS_UMASK, 0},
	{"getuid", syscall.SYS_GETUID, 0},
	{"getgid", syscall.SYS_GETGID, 0},
	{"vhangup", syscall.SYS_VHANGUP, 0},
	{"sync", syscall.SYS_SYNC, 0},
	{"gettid", syscall.SYS_GETTID, 0},
	{"time", syscall.SYS_TIME, ptr1},
	{"posix_fadvise", syscall.SYS_FADVISE64, 0},
	{"_exit", syscall.SYS_EXIT_GROUP, 0},
	{"inotify_init", syscall.SYS_INOTIFY_INIT, 0},
	{"inotify_add_watch", syscall.SYS_INOTIFY_ADD_WATCH, ptr2},
	{"inotify_rm_watch", syscall.SYS_INOTIFY_RM_WATCH, 0},
	{"openat", syscall.SYS_OPENAT, ptr2},
	{"mkdirat", syscall.SYS_MKDIRAT, ptr2},
	{"fchownat", syscall.SYS_FCHOWNAT, ptr2},
	{"unlinkat", syscall.SYS_UNLINKAT, ptr2},
	{"renameat", syscall.SYS_RENAMEAT, ptr2 | ptr4},
	{"linkat", syscall.SYS_LINKAT, ptr2 | ptr4},
	{"symlinkat", syscall.SYS_SYMLINKAT, ptr1 | ptr3},
	{"readlinkat", syscall.SYS_READLINKAT, ptr2 | ptr3},
	{"fchmodat", syscall.SYS_FCHMODAT, ptr2},
	{"faccessat", syscall.SYS_FACCESSAT, ptr2},
	{"splice", syscall.SYS_SPLICE, ptr2 | ptr4},
	{"tee", syscall.SYS_TEE, 0},
	{"sync_file_range", syscall.SYS_SYNC_FILE_RANGE, 0},
	{"fallocate", syscall.SYS_FALLOCATE, 0},
	{"eventfd", syscall.SYS_EVENTFD2, 0},
	{"dup3", syscall.SYS_DUP3, 0},
	{"pipe2", syscall.SYS_PIPE2, ptr1},
}
