@echo off
mkdir pack
copy fib.ecal pack
robocopy lib pack\lib
..\..\ecal.exe pack -dir pack -target out.exe fib.ecal
