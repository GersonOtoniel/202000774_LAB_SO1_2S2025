savedcmd_mod_sysinfo.mod := printf '%s\n'   mod_sysinfo.o | awk '!x[$$0]++ { print("./"$$0) }' > mod_sysinfo.mod
