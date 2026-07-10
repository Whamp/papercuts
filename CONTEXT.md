# Papercuts

Papercuts captures workflow friction that agents encounter while working so a user can improve the environment later.

## Language

**Papercut**:
A concrete instance of workflow friction encountered by a human or agent while pursuing another task.
_Avoid_: Issue, complaint

**Capture**:
The act of appending one papercut to its project or global log.
_Avoid_: Report, ticket creation

**Attribution**:
Optional, caller-supplied labels for the reporter and model associated with a capture. Papercuts never infer identity from Git, the operating-system account, or the environment.

**Project scope**:
The work contained by the agent session’s starting directory.
_Avoid_: Repository scope, Git scope

**Global scope**:
Work outside the current project or in tooling and environment shared across projects.
_Avoid_: Project scope

**Severity**:
The reporter’s required classification of a papercut as low, medium, or high impact. Low means an avoidable detour that did not change the approach or confidence in the result. Medium means meaningful rework, repeated attempts, a workaround, a changed approach, or reduced confidence while the task remained safely completable. High means blocked completion, required human or environment intervention, or credible risk of an incorrect, destructive, or insecure result. When several meanings apply, use the highest severity.
