# Architecture Guidance

- Prefer deep modules: a boundary should hide meaningfully more complexity than the interface it adds.
- Do not split or merge modules by file length alone.
- Avoid shallow wrappers, pass-through hooks, pass-through components, pass-through services, and speculative abstractions.
- Favor feature-local boundaries that reduce reader knowledge and keep schema or workflow changes local.
- PR review question: does this change reduce the number of places or files a reader must understand for one behavior?
