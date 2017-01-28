# VEBBEN TODO

* Test coverage (duh).
* Add more date & time formats; basically everything in the `time` package.
* Figure out whether there is any *real* difference btw `[]rune(s)` and norm.
* Cache date formats in a mutex so we check the most common first.
    * (If first misses but N hits, cache N as first for next check.)
    * (Do this separately for the different date formats.)
