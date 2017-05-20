# FSGC
The simple file system garbage collector
# Build
`go build`

# Run
A root is directory where collector should walk, defaults is current folder

`./PATH_TO_EXECUTABLE --root=/path/to/root_dir`


# Schedule
`crontab -l | { cat; echo "0 * * * * PATH_TO_EXECUTABLE --root=/path/to/root_dir"; } | crontab -`

# Test
```
cd fsgc/
go test
```
