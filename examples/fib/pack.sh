#!/bin/sh
mkdir pack
cp fib.ecal pack
cp -fR lib pack
../../ecal pack -dir pack fib.ecal
