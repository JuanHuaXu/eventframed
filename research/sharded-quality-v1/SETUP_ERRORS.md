# Setup And Coverage Notes

Preparation returned a live process handle before manifest.json existed. A
premature dependent run exited ENOENT without starting any Go command. The
original preparation handle was polled to verified successful completion;
the runner then started once. No measurements or gates were overwritten.
This is an orchestration error, not a scientific arm failure. Await dependent
preparation completion rather than treating a missing manifest as terminal.

An initial documentation patch missed exact context and made no file change;
the actual line was read before applying the narrow correction. No instruction
tree or production setting was modified to work around either error.
