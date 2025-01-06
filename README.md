## RU

A simple competitive programming problem parser and test runner.

Parsing Features

- [x] Problems
- [x] Contests

This wokrs smoothly in Codeforces and Atcoder.

Install

```bash
go install github.com/kunalsin9h/ru@latest
```

Config

```bash
ru config
```

Here add your C++ compile command (without `-o`)

Example:

```bash
# command
ru config

# input
Paste your c++ compile command: g++ -std=c++17  -O2
```

Parsing Problem

```bash
# first do
ru parse
# or `ru p`

# then click on Competitive Companion Browser Extension
```

```bash
# cd A
ru test
# or `ru t`
```

Copy source when test passes

```bash
ru test --copy
# or `ru t --copy`
# or `ru t -c`

# this need clipboard utility like xclip
```

Parsing Contests

When parsing contest we need to give the `number-or-problems` argument to the `parse` command

```bash
ru parse 6
# this will parse 6 problems from contest
```

