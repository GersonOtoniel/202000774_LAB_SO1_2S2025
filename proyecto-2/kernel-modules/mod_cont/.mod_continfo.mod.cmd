savedcmd_mod_continfo.mod := printf '%s\n'   mod_continfo.o | awk '!x[$$0]++ { print("./"$$0) }' > mod_continfo.mod
