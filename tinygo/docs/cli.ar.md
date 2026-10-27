# توثيق CLI

[English](cli.md) | **العربية**

الأداة `binarytable` مبنية من `cmd/binarytable` ويقع تنفيذها في `internal/cli`.
تستخدم المكتبة القياسية فقط.

```bash
go run ./cmd/binarytable help
go install ./cmd/binarytable
binarytable version
```

أكواد الخروج:

| الكود | المعنى |
| --- | --- |
| `0` | نجاح |
| `1` | خطأ تنفيذ أو بيانات غير صالحة أو "غير موجود" |
| `2` | خطأ في الاستخدام |

`-in -` و`-out -` تعنيان stdin وstdout في الأوامر التي تدعم ذلك. عند كتابة
بيانات ثنائية إلى stdout تُكتب رسائل الحالة إلى stderr.

---

## 1. أوامر الجدول

```text
binarytable table <create|info|schema|get|set|dump|search> [flags]
```

### مواصفات الأعمدة

الخيار `-col` يقبل `name:type[:length]` ويمكن تكراره.

| النوع | المعنى | الطول |
| --- | --- | --- |
| `int`, `int32` | عدد صحيح 4 بايت | 4 أو يُحذف |
| `long`, `int64` | عدد صحيح 8 بايت | 8 أو يُحذف |
| `double`, `float64` | عدد عشري 8 بايت | 8 أو يُحذف |
| `string`, `str`, `text` | نص UTF-8 | افتراضي 64 |
| `raw`, `bytes`, `blob` | بايتات خامة | مطلوب |

```bash
binarytable table create -file users.db -rows 1000 \
  -col _id:int -col name:string:64 -col score:double -col payload:raw:1024
```

### `table create`

| الخيار | مطلوب | الوصف |
| --- | --- | --- |
| `-file` | نعم | ملف الخرج |
| `-rows` | نعم | سعة الصفوف |
| `-col` | نعم | مواصفة عمود، قابلة للتكرار |
| `-force` | لا | استبدال ملف موجود |

`CreateFile` يقتطع الملف، لذلك ترفض الأداة الاستبدال بدون `-force`.

### `table info`

```bash
binarytable table info -file users.db
```

يطبع حجم الملف، عدد الصفوف، حجم الصف، حجم المخطط، السعة، والأعمدة.

### `table schema`

```bash
binarytable table schema -file users.db
```

يطبع نص المخطط الخام (`name/length/TYPE\n`).

### `table get`

| الخيار | الوصف |
| --- | --- |
| `-file` | ملف الجدول |
| `-row` | فهرس الصف |
| `-col` | اسم العمود أو فهرسه |
| `-hex` | طباعة البايتات الخام بصيغة hex |
| `-out` | كتابة البايتات الخام إلى ملف، أو `-` لـ stdout |

```bash
binarytable table get -file users.db -row 0 -col name
binarytable table get -file users.db -row 0 -col payload -out payload.bin
binarytable table get -file users.db -row 0 -col payload -hex
```

أعمدة RAW تُطبع بصيغة hex افتراضياً. لا يمكن الجمع بين `-hex` و`-out`.

### `table set`

| الخيار | الوصف |
| --- | --- |
| `-file` | ملف الجدول |
| `-row` | فهرس الصف |
| `-col` | اسم العمود أو فهرسه |
| `-value` | قيمة مكتوبة: نص، int، long، double، أو نص خام |
| `-value-hex` | بايتات خامة بصيغة hex |
| `-value-file` | قراءة بايتات خامة من ملف، أو `-` من stdin |

يجب تحديد خيار قيمة واحد فقط.

```bash
binarytable table set -file users.db -row 0 -col _id -value 1
binarytable table set -file users.db -row 0 -col name -value Alice
binarytable table set -file users.db -row 0 -col payload -value-hex 010203
binarytable table set -file users.db -row 0 -col payload -value-file avatar.png
```

`-value-hex` و`-value-file` يكتبان بايتات خامة. لعمود `STRING` استخدم `-value`
حتى تُصفَّر بقية العمود.

### `table dump`

```bash
binarytable table dump -file users.db -row 0
```

الخرج مفصول بعلامات جدولة: `index`, `name`, `TYPE`, `value`. النصوص بين
علامتي اقتباس، والقيم الخام بصيغة hex.

### `table search`

| الخيار | الوصف |
| --- | --- |
| `-file` | ملف الجدول |
| `-col` | اسم العمود أو فهرسه |
| `-value` | القيمة المطلوبة |
| `-start` | أول صف يُفحص (افتراضي 0) |

```bash
binarytable table search -file users.db -col name -value Alice
```

يطبع فهرس الصف، أو `-1` ويخرج بالكود 1 عند عدم العثور عليه. البحث في RAW غير
مدعوم.

---

## 2. أوامر قائمة البلوبات

```text
binarytable list <create|info|add|get|dump> [flags]
```

### `list create`

```bash
binarytable list create -file blobs.bin -max 5000
```

- `-max` يجب أن يكون موجباً ويُحفظ بحجز `100 + maxEntries × 4` بايت.
- `-force` يستبدل ملفاً موجوداً.
- البادئة المحجوزة تسمح لـ `list info` و`list add` باستعادة السعة تلقائياً حتى
  لو كانت القائمة فارغة.

### `list info`

```bash
binarytable list info -file blobs.bin
```

`-max 0` (الافتراضي) يستعيد السعة من البادئة المحجوزة أو من أول إزاحة بيانات.
مرّر `-max N` صراحة للملفات القديمة التي يكون كشف سعتها ملتبساً.

### `list add`

| الخيار | الوصف |
| --- | --- |
| `-file` | ملف القائمة (يُنشأ إذا لم يوجد) |
| `-in` | ملف البلوب، أو `-` لـ stdin |
| `-max` | السعة؛ `0` يستعيدها، والملفات الجديدة افتراضياً 1000 |

```bash
binarytable list add -file blobs.bin -in avatar.png
printf 'hello' | binarytable list add -file blobs.bin -in -
```

### `list get`

| الخيار | الوصف |
| --- | --- |
| `-file` | ملف القائمة |
| `-index` | فهرس البلوب |
| `-max` | سعة القائمة |
| `-hex` | طباعة hex |
| `-out` | كتابة إلى ملف، أو `-` لـ stdout |

```bash
binarytable list get -file blobs.bin -index 0 -out avatar-copy.png
binarytable list get -file blobs.bin -index 0 -hex
```

### `list dump`

```bash
binarytable list dump -file blobs.bin
```

يطبع `index<TAB>size` لكل بلوب.

---

## 3. أوامر الضغط

### `deflate`

```text
binarytable deflate <pack|write|unpack|read|info> [flags]
```

`pack` / `write`:

| الخيار | الافتراضي | الوصف |
| --- | --- | --- |
| `-in` | مطلوب | ملف الدخل، أو `-` لـ stdin |
| `-out` | مطلوب | ملف الخرج، أو `-` لـ stdout |
| `-block` | 8192 | حجم الكتلة |
| `-header` | 0 | بايتات رأس محجوزة للمتصل |
| `-force` | false | استبدال ملف الخرج |

`unpack` / `read`:

| الخيار | الافتراضي | الوصف |
| --- | --- | --- |
| `-in` | مطلوب | ملف الدخل، أو `-` لـ stdin |
| `-out` | مطلوب | ملف الخرج، أو `-` لـ stdout |
| `-cache` | 8 | عدد الكتل المفكوكة في الكاش |
| `-force` | false | استبدال ملف الخرج |

`info` يقرأ رأس الكتل ويطبع الطول الأصلي والمضغوط والنسبة وحجم الكتلة ورأس
المتصل وعدد الكتل:

```bash
binarytable deflate pack -in data.bin -out data.deflate -block 8192
binarytable deflate unpack -in data.deflate -out data.bin -cache 4
binarytable deflate info -in data.deflate

# قراءة صريحة بنمط DeflateBinaryStreamReader وتمريرها لأداة أخرى:
binarytable deflate read -in data.deflate -out - | sha256sum
```

`read` اسم بديل لـ `unpack`، و`write` اسم بديل لـ `pack`.

### `zlib`

```text
binarytable zlib <compress|decompress> -in PATH -out PATH
```

```bash
binarytable zlib compress -in data.bin -out data.zlib
binarytable zlib decompress -in data.zlib -out data.bin
```

كلا الأمرين يقبل `-` لـ stdin/stdout.

---

## 4. أوامر AES

```text
binarytable aes <iv|encrypt|decrypt> [flags]
```

`aes iv` يطبع IV عشوائياً بطول 16 بايت بصيغة hex.

`aes encrypt` / `aes decrypt`:

| الخيار | الوصف |
| --- | --- |
| `-in` | ملف الدخل، أو `-` لـ stdin |
| `-out` | ملف الخرج، أو `-` لـ stdout |
| `-password` | كلمة المرور (ظاهرة في وسائط العملية) |
| `-password-file` | ملف يحتوي كلمة المرور |
| `-iv` | IV بطول 16 بايت بصيغة hex، مطلوب |

يجب تحديد واحد فقط من `-password` و`-password-file`. يُفضّل `-password-file`
لأن وسائط سطر الأوامر قد تكون مرئية لعمليات أخرى.

```bash
binarytable aes iv
binarytable aes encrypt -in plain.bin -out secret.bin \
  -password-file password.txt -iv 000102030405060708090a0b0c0d0e0f
binarytable aes decrypt -in secret.bin -out plain.bin \
  -password-file password.txt -iv 000102030405060708090a0b0c0d0e0f
```

---

## 5. أوامر RSA

```text
binarytable rsa <generate|sign|verify|encrypt|decrypt> [flags]
```

| الأمر | الخيارات |
| --- | --- |
| `generate` | `-bits` (افتراضي 2048)، `-public`، `-private` |
| `sign` | `-key`، `-in`، `-out` |
| `verify` | `-key`، `-in`، `-sig` |
| `encrypt` | `-key`، `-in`، `-out` |
| `decrypt` | `-key`، `-in`، `-out` |

```bash
binarytable rsa generate -bits 2048 -public public.pem -private private.pem
binarytable rsa sign   -key private.pem -in message.bin -out message.sig
binarytable rsa verify -key public.pem  -in message.bin -sig message.sig
```

`verify` يطبع `signature valid` ويخرج بالكود 0، أو يخرج بالكود 1 مع
`signature invalid`.

RSA مخصص للأهداف المستضافة؛ عادةً يكون ثقيلاً على المتحكمات الصغيرة.

---

## 6. أمثلة shell

إنشاء جدول وتعبئته وفحصه:

```bash
set -e

binarytable table create -file users.db -rows 100 -force \
  -col _id:int -col name:string:32 -col score:double

binarytable table set -file users.db -row 0 -col _id -value 1
binarytable table set -file users.db -row 0 -col name -value Alice
binarytable table set -file users.db -row 0 -col score -value 9.5

binarytable table info -file users.db
binarytable table dump -file users.db -row 0
```

ضغط ملف جدول:

```bash
binarytable deflate pack -in users.db -out users.db.deflate
binarytable deflate unpack -in users.db.deflate -out users.db.copy
```
