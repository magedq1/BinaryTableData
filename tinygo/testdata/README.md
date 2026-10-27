# Kotlin compatibility fixtures

This directory contains binary files produced by the **original Kotlin
implementation**. The Go tests read them to prove that the port understands the
same on-disk format.

| File | Produced by | Read by |
| --- | --- | --- |
| `table.bin` | `BinaryTableData` over `ByteArrayStreamAutoSize` | `TestKotlinTableFixture` |
| `table-file.bin` | `BinaryTableData(header, File)` | `TestKotlinTableFileFixture` |
| `list.bin` | `BinaryList` | `TestKotlinBinaryListFixture` |
| `deflate.bin` | `reader.deflate(writer, 1024, 0)` | `TestKotlinDeflateFixture` |
| `deflate-raw.bin` | original data for `deflate.bin` | `TestKotlinDeflateFixture` |
| `aes.bin` | `AESEncryptor.encrypt` | `TestKotlinAESFixture` |
| `aes-iv.bin` | fixed IV `00 01 ... 0f` | `TestKotlinAESFixture` |
| `aes-plain.bin` | plaintext | `TestKotlinAESFixture` |

## Regenerating the fixtures

The generator uses the compiled Kotlin classes from the repository:

- `../lib/build/classes/kotlin/main`
- the Kotlin 1.9.23 standard library from the Gradle cache.

### PowerShell (Windows)

```powershell
$repo = "D:\workspace\BinaryTableData"
$classes = "$repo\lib\build\classes\kotlin\main"
$stdlib = Get-ChildItem "$env:USERPROFILE\.gradle\caches\modules-2\files-2.1\org.jetbrains.kotlin\kotlin-stdlib\1.9.23" -Recurse -Filter kotlin-stdlib-1.9.23.jar | Select-Object -First 1 -ExpandProperty FullName
$cp = "$classes;$stdlib"

javac -cp $cp -d "$env:TEMP\btable-gen" kotlin\GenerateFixtures.java
java -cp "$cp;$env:TEMP\btable-gen" GenerateFixtures .
```

### POSIX shell

```bash
repo=../..
classes="$repo/lib/build/classes/kotlin/main"
stdlib=$(find "$HOME/.gradle/caches/modules-2/files-2.1/org.jetbrains.kotlin/kotlin-stdlib/1.9.23" -name 'kotlin-stdlib-1.9.23.jar' | head -n 1)
cp="$classes:$stdlib"

javac -cp "$cp" -d /tmp/btable-gen kotlin/GenerateFixtures.java
java -cp "$cp:/tmp/btable-gen" GenerateFixtures .
```

## Reverse check (Go → Kotlin)

`kotlin/VerifyGoFixtures.java` reads files written by the Go port and checks
them with the Kotlin library. Generate the Go files from the `tinygo` directory:

```bash
go run ./testdata/kotlin/write_go_fixtures.go /tmp/go-fixtures
```

Then run:

```bash
javac -cp "$cp" -d /tmp/btable-gen kotlin/VerifyGoFixtures.java
java -cp "$cp:/tmp/btable-gen" VerifyGoFixtures /tmp/go-fixtures
```

The reverse check was executed during development and passed for tables,
files, `BinaryList`, and Deflate streams. AES is byte-for-byte deterministic and
is checked directly by the Go test.
