# How to add a new job?
1. See `types/jobs.go`, add a new job category there. Be descriptive. Do not change the string values ever. Not worth the hassle.
2. In `jobs/receivers.go`, add the category to `catmap` and map it to a receiver function which shall lie in the same file.
3. In `jobs/senders.go` write a sender of the job. It should be public. Use `planJob` function. They probably will receive the reports from the `activities` package.