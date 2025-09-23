#include <linux/module.h>
#include <linux/init.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/mm.h>
#include <linux/sched/signal.h>

#define PROC_NAME "sysinfo_so1_202000774"

static int sysinfo_show(struct seq_file *m, void *v) {
    struct sysinfo si;
    si_meminfo(&si);
    unsigned long total_mb = (si.totalram * si.mem_unit) / (1024*1024);
    unsigned long free_mb = (si.freeram * si.mem_unit) / (1024*1024);
    unsigned long used_mb = total_mb - free_mb;
    seq_printf(m, "Total_RAM_MB: %lu\nFree_RAM_MB: %lu\nUsed_RAM_MB: %lu\n", total_mb, free_mb, used_mb);

    seq_printf(m, "\n--- Procesos del sistema ---\n");
    {
        struct task_struct *task;
        for_each_process(task) {
            unsigned long vsize_mb = 0;
            long rss_mb = 0;
            if (task->mm) {
                vsize_mb = (task->mm->total_vm << (PAGE_SHIFT - 10));
                rss_mb = get_mm_rss(task->mm) << (PAGE_SHIFT - 10);
            }
            char state = task_state_to_char(task);
            seq_printf(m, "PID:%d Name:%s CMD:%s VSZ_MB:%lu RSS_MB:%ld State:%c\n",
                task->pid, task->comm, task->comm, vsize_mb, rss_mb, state);
        }
    }

    return 0;
}

static int sysinfo_open(struct inode *inode, struct file *file) {
    return single_open(file, sysinfo_show, NULL);
}


static const struct proc_ops sysinfo_ops = {
    .proc_open    = sysinfo_open,
    .proc_read    = seq_read,
    .proc_lseek   = seq_lseek,
    .proc_release = single_release,
};

static int __init sysinfo_init(void) {
    proc_create(PROC_NAME, 0444, NULL, &sysinfo_ops);
    pr_info("sysinfo module loaded, /proc/%s created\n", PROC_NAME);
    return 0;
}

static void __exit sysinfo_exit(void) {
    remove_proc_entry(PROC_NAME, NULL);
    pr_info("sysinfo module removed\n");
}

MODULE_LICENSE("GPL");
MODULE_AUTHOR("Gerson");
MODULE_DESCRIPTION("Modulo /proc para info del sistema");
module_init(sysinfo_init);
module_exit(sysinfo_exit);

