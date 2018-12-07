# JarDB

Quickly find JARs providing a certain Java class

In the past I got in the situation from time to time that I needed a specific Java class for some task and I didn't
know which Java library provided that class. Enter JarDB: It lets you build an index of the JAR files on your computer (e.g.
in the `$HOME/.m2/` folder) and all the classes they provide:

```shell
$ jardb index ~/.m2
```
The command above indexes all JARs and their Java classes for a fast lookup like this:

```shell
$ jardb find com.sun.star.lib.unoloader.UnoLoader
```
You can use JarDB to find a class by its fully qualified name or by parts of the name (e.g. `UnoLoader`). To make JarDB more
comfortable to use, you can also provide the class you're searching for as a path like in
`com/sun/star/lib/unoloader/UnoLoader.class`.

# Installation

If you have a Go installation on your computer just issue `go get github.com/makkes/jardb` and you're ready to go.

Otherwise, grab the [most current release](https://github.com/makkes/jardb/releases).
