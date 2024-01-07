# How to add a new job?
1. See `jobtype.go`, add a new job category there. Be descriptive. Do not change the string values ever. Not worth the hassle.
2. In `jobs/receivers.go`, add the category to `catmap` and map it to a receiver function which shall lie in the same file.
3. Use functions `ScheduleJSON` and `ScheduleDatum` to schedule jobs.

If your job is not making any expensive operations such as network requests or many database requests, then you probably should not make a job.