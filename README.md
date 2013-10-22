dircomp
=======

Command line utility to list file differences in two directory trees

    Usage: dircomp [flags] sourceDirectory destinationDirectory

    Compare source directory tree to destination, listing differences
    in destination, based on an MD5 hash comparison.

    Output indicators:

    + file is added in destination
    - file is removed in destination
    M file is modified in destination
      file is identical (only with -all flag)

    -all=false: show all files, not just changes
    -debug=false: print debugging information
    -help=false: print this help message
