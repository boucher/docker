package libcontainer

type CriuOpts struct {
    ImagesDirectory             string // directory for storing image files
    WorkDirectory               string // directory to cd and write logs/pidfiles/stats to
    PreviousImagesDirectory     string // path to images from previous dump (relative to --images-directory)
    LogFile                     string // log file name
    LeaveRunning                bool   // leave container in running state after checkpoint
    LeaveStopped                bool   // leave container in stopped state after checkpoint
    TcpEstablished              bool   // checkpoint/restore established TCP connections
    ExternalUnixConnections     bool   // allow external unix connections
    ShellJob                    bool   // allow to dump and restore shell jobs
}

/* CRIU help:

Commands:
  dump           checkpoint a process/tree identified by pid
  pre-dump       pre-dump task(s) minimizing their frozen time
  restore        restore a process/tree
  show           show dump file(s) contents
  check          checks whether the kernel support is up-to-date
  exec           execute a system call by other task
  page-server    launch page server
  service        launch service
  dedup          remove duplicates in memory dump
  cpuinfo dump   writes cpu information into image file
  cpuinfo check  validates cpu information read from image file

Dump/Restore options:

* Generic:
  -t|--tree PID         checkpoint a process tree identified by PID
  -d|--restore-detached detach after restore
  -S|--restore-sibling  restore root task as sibling
  -s|--leave-stopped    leave tasks in stopped state after checkpoint
  -R|--leave-running    leave tasks in running state after checkpoint
  -D|--images-dir DIR   directory for image files
     --pidfile FILE     write root task, service or page-server pid to FILE
  -W|--work-dir DIR     directory to cd and write logs/pidfiles/stats to
                        (if not specified, value of --images-dir is used)
     --cpu-cap [CAP]    require certain cpu capability. CAP: may be one of:
                        'cpu','fpu','all','ins','none'. To disable capability, prefix it with '^'.
     --exec-cmd         execute the command specified after '--' on successful
                        restore making it the parent of the restored process

* Special resources support:
  -x|--ext-unix-sk      allow external unix connections
     --tcp-established  checkpoint/restore established TCP connections
  -r|--root PATH        change the root filesystem (when run in mount namespace)
  --evasive-devices     use any path to a device file if the original one
                        is inaccessible
  --veth-pair IN=OUT    map inside veth device name to outside one
                        can optionally append @<bridge-name> to OUT for moving
                        the outside veth to the named bridge
  --link-remap          allow to link unlinked files back when possible
  --action-script FILE  add an external action script
  -j|--shell-job        allow to dump and restore shell jobs
  -l|--file-locks       handle file locks, for safety, only used for container
  -L|--libdir           path to a plugin directory (by default /var/lib/criu/)
  --force-irmap         force resolving names for inotify/fsnotify watches
  -M|--ext-mount-map KEY:VALUE
                        add external mount mapping
  --manage-cgroups      dump or restore cgroups the process is in
  --cgroup-root [controller:]/newroot
                        change the root cgroup the controller will be
                        installed into. No controller means that root is the
                        default for all controllers not specified.

* Logging:
  -o|--log-file FILE    log file name
     --log-pid          enable per-process logging to separate FILE.pid files
  -v[NUM]               set logging level (higher level means more output):
                          -v1|-v    - only errors and messages
                          -v2|-vv   - also warnings (default level)
                          -v3|-vvv  - also information messages and timestamps
                          -v4|-vvvv - lots of debug

* Memory dumping options:
  --track-mem           turn on memory changes tracker in kernel
  --prev-images-dir DIR path to images from previous dump (relative to -D)
  --page-server         send pages to page server (see options below as well)
  --auto-dedup          when used on dump it will deduplicate "old" data in
                        pages images of previous dump
                        when used on restore, as soon as page is restored, it
                        will be punched from the image.

Page/Service server options:
  --address ADDR        address of server or service
  --port PORT           port of page server
  -d|--daemon           run in the background after creating socket

Show options:
  -f|--file FILE        show contents of a checkpoint file
  -F|--fields FIELDS    show specified fields (comma separated)
  -D|--images-dir DIR   directory where to get images from
  -c|--contents         show contents of pages dumped in hexdump format
  -p|--pid PID          show files relevant to PID (filter -D flood)

Other options:
  -h|--help             show this text
  -V|--version          show version
     --ms               don't check not yet merged kernel features
*/
