# BrainFrickAOT
A BF ahead of time interpreter and compiler which transpiles and optimizes BF into Golang and then interprets or compiles the end result

The basic workflow is:
```
Brainfrick -> AST -> Golang -> Optimized Golang -> Interpretation from Golpal | Compile and run golang
```

Outputted binaries can also be shrinked with UPX (https://upx.github.io/)
