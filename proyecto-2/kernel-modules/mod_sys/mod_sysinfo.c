#include <linux/module.h>
#include <linux/init.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/mm.h>
#include <linux/sched/signal.h>
#include <linux/sched/cputime.h>
#include <linux/jiffies.h>

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
        unsigned long total_jiffies = jiffies;
        for_each_process(task) {
            unsigned long vsize_mb = 0;
            long rss_mb = 0;
            if (task->mm) {
                vsize_mb = (task->mm->total_vm << (PAGE_SHIFT - 10)) / 1024;  
                rss_mb = (get_mm_rss(task->mm) << (PAGE_SHIFT - 10)) / 1024;
            }
          
            char state = task_state_to_char(task);

            unsigned long utime = task->utime;
            unsigned long stime = task->stime;
        
            unsigned long utime_sec = utime / HZ;
            unsigned long stime_sec = stime / HZ;

            unsigned long total_time = utime + stime;
            unsigned long cpu_usage = 0;
            if(total_jiffies > 0){
              cpu_usage = (total_time*100) / total_jiffies;
              cpu_usage /= num_online_cpus();
            }
            seq_printf(m, "PID:%d Name:%s CMD:%s VSZ_MB:%lu RSS_MB:%ld State:%c CPU_user:%lus CPU_system:%lus CPU_usage:%lu%%\n",
                task->pid, task->comm, task->comm, vsize_mb, rss_mb, state, utime_sec, stime_sec, cpu_usage );
        }
    }
    seq_printf(m, "\n");
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

