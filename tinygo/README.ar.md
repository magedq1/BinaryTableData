# BinaryTableData — نسخة Go / TinyGo

[English](README.md) | **العربية**

هذه نسخة Go موثّقة من مكتبة **BinaryTableData** المكتوبة أصلاً بلغة Kotlin.
تحافظ النسخة على **نفس صيغة الملفات الثنائية** (Big-endian) وتضيف واجهة مناسبة
لـ Go وTinyGo، مع اختبارات توافق تقرأ ملفات أنشأتها المكتبة الأصلية وتتحقق من
قراءة Kotlin للملفات التي ينتجها منفذ Go.

- الحزمة الأساسية: `github.com/magedq1/BinaryTableData/tinygo/binarytable`
- حزمة التشفير الاختيارية: `github.com/magedq1/BinaryTableData/tinygo/cryptoutil`
- أداة سطر الأوامر: `github.com/magedq1/BinaryTableData/tinygo/cmd/binarytable`
- المكتبة الأساسية وCLI مبنيان على المكتبة القياسية فقط. الواجهة الرسومية
  الاختيارية لها اعتماديات Wails/React الخاصة بها.

---

## المحتويات

1. [التوافق مع الإصدار الأصلي](#التوافق-مع-الإصدار-الأصلي)
2. [المتطلبات والتثبيت](#المتطلبات-والتثبيت)
3. [بدء سريع](#بدء-سريع)
4. [أداة سطر الأوامر](#أداة-سطر-الأوامر)
5. [الواجهة الرسومية Wails](#الواجهة-الرسومية-wails)
6. [بنية المشروع](#بنية-المشروع)
7. [التوثيق التفصيلي](#التوثيق-التفصيلي)
8. [ملاحظات TinyGo](#ملاحظات-tinygo)
9. [الاختبارات](#الاختبارات)
10. [الفروق المقصودة عن Kotlin](#الفروق-المقصودة-عن-kotlin)

---

## التوافق مع الإصدار الأصلي

تم التحقق من التوافق في الاتجاهين باستخدام مكتبة Kotlin المترجمة نفسها:

| العنصر | Kotlin يكتب → Go يقرأ | Go يكتب → Kotlin يقرأ |
| --- | --- | --- |
| جدول `BinaryTableData` في الذاكرة | ✅ | ✅ |
| جدول `BinaryTableData` في ملف | ✅ | ✅ |
| `BinaryList` | ✅ | ✅ |
| ضغط Deflate ذو الكتل | ✅ | ✅ |
| AES/CBC/PKCS#5 | ✅ (بايتات مطابقة) | ✅ (بايتات مطابقة) |

الملفات الثنائية الموجودة في `testdata/` مولّدة بواسطة Kotlin، واختبارات
`*_compat_test.go` تقرأها. سكربتات Java في `testdata/kotlin/` توثّق كيفية
إعادة توليدها والتحقق من الاتجاه المعاكس.

---

## المتطلبات والتثبيت

- **Go 1.22 أو أحدث** للمكتبة وCLI.
- **TinyGo حديث** (يُنصح بأحدث إصدار) عند استهداف المتحكمات أو WASI.
- **Go 1.23+ وWails v2.17+ وNode.js 20+ وWebView2** للواجهة الرسومية.
- المكتبة لا تحتاج CGO ولا اعتماديات خارجية؛ وحدة الواجهة وحدها تجلب Wails
  وحزم الواجهة الأمامية.

المشروع داخل مستودع Kotlin الأصلي. يمكن استخدامه مباشرة عبر `replace`:

```bash
go mod init example.com/myapp
go mod edit -replace github.com/magedq1/BinaryTableData/tinygo=../BinaryTableData/tinygo
go get github.com/magedq1/BinaryTableData/tinygo
```

أو من داخل مجلد `tinygo/`:

```bash
go test ./...
go run ./examples/demo
go run ./cmd/binarytable help
```

---

## بدء سريع

```go
package main

import (
	"fmt"
	"log"

	"github.com/magedq1/BinaryTableData/tinygo/binarytable"
)

func main() {
	header, err := binarytable.NewHeader(1000, []binarytable.Column{
		binarytable.IntColumn("_id"),
		binarytable.StringColumn("first_name", 64),
		binarytable.StringColumn("last_name", 64),
	})
	if err != nil {
		log.Fatal(err)
	}

	// جدول في الذاكرة. استخدم CreateFile لجدول ملفات ثابت السعة.
	stream := binarytable.NewAutoSizeStream(1024)
	table, err := binarytable.New(header, stream)
	if err != nil {
		log.Fatal(err)
	}

	if err := table.SetInt(0, 0, 1); err != nil {
		log.Fatal(err)
	}
	if err := table.SetString(0, 1, "Alice"); err != nil {
		log.Fatal(err)
	}
	if err := table.SetStringByName(0, "last_name", "Saleh"); err != nil {
		log.Fatal(err)
	}

	name, err := table.GetStringByName(0, "first_name")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(name) // Alice

	// إعادة فتح الجدول من البايتات (للقراءة فقط).
	reopened, err := binarytable.FromBytes(stream.ToBytes())
	if err != nil {
		log.Fatal(err)
	}
	_ = reopened
}
```

مثال كامل قابل للتشغيل: [`examples/demo/main.go`](examples/demo/main.go).

---

## أداة سطر الأوامر

الأداة `binarytable` تغطي الجداول والقوائم والضغط والتشفير:

```bash
go install ./cmd/binarytable
binarytable help
binarytable table help
```

إنشاء جدول واستخدامه:

```bash
binarytable table create -file users.db -rows 1000 \
  -col _id:int -col name:string:64 -col score:double

binarytable table set -file users.db -row 0 -col _id -value 1
binarytable table set -file users.db -row 0 -col name -value Alice
binarytable table get -file users.db -row 0 -col name
binarytable table dump -file users.db -row 0
binarytable table search -file users.db -col name -value Alice
binarytable table info -file users.db
```

قوائم البلوبات:

```bash
binarytable list create -file blobs.bin -max 5000
binarytable list add -file blobs.bin -in avatar.png
binarytable list get -file blobs.bin -index 0 -out avatar-copy.png
binarytable list dump -file blobs.bin
```

`list info` و`list add` يستعيدان السعة تلقائياً للقوائم التي ينشئها هذا المنفذ؛
استخدم `-max N` للملفات القديمة أو الملتبسة.

الضغط:

```bash
binarytable zlib compress -in data.bin -out data.zlib
binarytable zlib decompress -in data.zlib -out data.bin

binarytable deflate pack -in data.bin -out data.deflate -block 8192
binarytable deflate unpack -in data.deflate -out data.bin
```

التشفير:

```bash
binarytable aes iv
binarytable aes encrypt -in plain.bin -out secret.bin \
  -password-file password.txt -iv 000102030405060708090a0b0c0d0e0f

binarytable rsa generate -bits 2048 -public public.pem -private private.pem
binarytable rsa sign   -key private.pem -in message.bin -out message.sig
binarytable rsa verify -key public.pem  -in message.bin -sig message.sig
```

`-in -` و`-out -` تعنيان stdin وstdout في الأوامر التي تدعم ذلك. عند كتابة
بيانات ثنائية إلى stdout تُكتب رسائل الحالة إلى stderr. الأمر `deflate info`
يقرأ رأس الكتل، و`deflate read` اسم صريح لعملية فك الضغط المكافئة لـ
`DeflateBinaryStreamReader`.

راجع [توثيق CLI](docs/cli.ar.md) لكل الأوامر والخيارات.

---

## الواجهة الرسومية Wails

الوحدة `gui/` هي محرر سطح مكتب كامل مبني بـ Wails v2 وReact 19 وTailwind CSS 4:

- منشئ مخطط الجدول، شبكة بيانات بترقيم صفحات، مفتش خلايا حسب النوع، بحث،
  مسح صفوف، وتضمين ملفات في خلايا RAW.
- محرر قوائم بلوبات بسعة صريحة، تضمين متعدد بالبث، معاينة صور/نصوص/hex،
  تصدير، استبدال، وحذف ذري يعيد بناء القائمة.
- مختبر Deflate للضغط وفك الضغط وفحص الرأس وضبط الكاش، مع **تصفح للقراءة فقط
  للجداول/القوائم المضغوطة عبر `DeflateReader`**.
- إفلات عام للملفات، حوارات أصلية، لوحة أوامر (`Ctrl+K`)، ونظام تصميم Dark OLED.

```bash
cd tinygo/gui
go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.0
wails dev       # إعادة تحميل حرارية
wails build -clean
```

يُكتب الملف التنفيذي إلى `build/bin/BinaryTableStudio.exe`.
راجع [توثيق الواجهة](gui/README.ar.md) للمزايا والاختصارات ونظام التصميم.

---

## بنية المشروع

```
tinygo/
├── go.mod
├── README.md / README.ar.md
├── binarytable/          الحزمة الأساسية
│   ├── stream.go         Reader / Writer / Stream / Grower
│   ├── bytearray.go      تدفقات الذاكرة
│   ├── file.go           تدفق الملفات (ReadAt/WriteAt)
│   ├── subreader.go      SubReader و JoinedReader
│   ├── column.go         الأعمدة وصيغة المخطط
│   ├── header.go         رأس الجدول
│   ├── table.go          الجدول والعمليات المكتوبة
│   ├── search.go         indexOf
│   ├── binarylist.go     قائمة البلوبات متغيرة الطول
│   ├── compression.go    zlib مباشر
│   ├── deflate.go        صيغة Deflate ذات الكتل
│   └── cache.go          LRU cache
├── cryptoutil/           AES و RSA و random
├── cmd/binarytable/      نقطة دخول CLI
├── internal/cli/         تنفيذ CLI واختباراته
├── gui/                  محرر Wails لسطح المكتب (Go + React)
├── examples/demo/        مثال تشغيلي
├── docs/                 التوثيق (الإنجليزية افتراضياً، العربية بلاحقة .ar.md)
└── testdata/             ملفات توليدها Kotlin + أدوات التحقق
```

---

## التوثيق التفصيلي

- [صيغة البيانات والتوافق](docs/format.ar.md)
- [الجدول والرأس والأعمدة](docs/core.ar.md)
- [التدفقات والقراءة العشوائية](docs/streams.ar.md)
- [قوائم البلوبات BinaryList](docs/binarylist.ar.md)
- [الضغط Deflate](docs/compression.ar.md)
- [الأدوات المساعدة: البحث، الترميز، الكاش](docs/helpers.ar.md)
- [التشفير AES / RSA](docs/security.ar.md)
- [توثيق CLI](docs/cli.ar.md)
- [الواجهة الرسومية Wails](gui/README.ar.md)
- [TinyGo: القيود والاستهداف](docs/tinygo.ar.md)
- [الانتقال من Kotlin إلى Go](docs/migration.ar.md)

النسخ الإنجليزية هي الافتراضية وبجانبها نسخ عربية باللاحقة `.ar.md`، مثل
[`docs/core.md`](docs/core.md) و[`docs/core.ar.md`](docs/core.ar.md).

---

## ملاحظات TinyGo

- **لا يوجد Serialization تلقائي بالانعكاس.** الإصدار الأصلي يستخدم
  `kotlinx.serialization`، بينما TinyGo يدعم `reflect` بشكل جزئي فقط. لذلك
  الواجهة هنا **schema-first**: عرّف الأعمدة صراحةً ثم اقرأ/اكتب الحقول.
- **تدفقات الذاكرة تعمل على المتحكمات.** `NewAutoSizeStream` و
  `NewByteArrayStream` و`ByteArrayReader` لا تعتمد على نظام ملفات.
- **تدفقات الملفات تحتاج هدفاً يستضيف نظام ملفات.** `OpenFile` و`CreateFile`
  تعمل على Linux وWindows وWASI وغيرها، وليست مناسبة لهدف bare-metal.
- **الضغط** يستخدم `compress/zlib` المدعوم في TinyGo.
- **التشفير AES** يستخدم `crypto/aes` و`crypto/cipher` و`crypto/sha256`.
  حزمة `cryptoutil` تشمل RSA أيضاً، لكنها اختيارية وثقيلة على المتحكمات.
- **واجهة Wails مخصصة لسطح المكتب** وليست هدف TinyGo. تبقى المكتبة الأساسية
  متوافقة مع TinyGo، والواجهة تطبيق إضافي على الأجهزة المستضيفة.
- لبناء مثال WASI:

  ```bash
  tinygo build -target=wasi -o demo.wasm ./examples/demo
  ```

- لاختبار الحزم باستخدام TinyGo (حسب الهدف المتاح):

  ```bash
  tinygo test ./...
  ```

---

## الاختبارات

```bash
cd tinygo
gofmt -l .
go vet ./...
go test ./...
go test -race ./...   # اختياري، يتطلب CGO
```

`go test ./...` يشغّل:

- اختبارات التدفقات (ذاكرة، ملفات، Sub/Joined).
- اختبارات الجدول والرأس والبحث.
- اختبارات `BinaryList`.
- اختبارات ضغط Deflate عبر حدود الكتل.
- اختبارات AES/RSA.
- اختبارات CLI (جداول، قوائم، ضغط، تشفير).
- اختبارات خدمات الواجهة الرسومية.
- اختبارات التوافق مع ملفات Kotlin الجاهزة.

للواجهة:

```bash
cd gui
go test ./...
cd frontend && npm run build
```

---

## الفروق المقصودة عن Kotlin

1. **الأخطاء بدل الاستثناءات**: الدوال تعيد `error`، ويمكن فحصها باستخدام
   `errors.Is(err, binarytable.ErrOverSize)` وغيرها.
2. **SetString يصفّر بقية العمود** لتجنّب بقايا نص أطول عند الكتابة فوقه.
   `SetValue` للبايتات الخام يبقى مطابقاً لسلوك Kotlin (لا يكتب أصفاراً).
3. **لا mmap**: تدفق الملفات يستخدم `ReadAt`/`WriteAt` بدل `MappedByteBuffer`.
4. **القراءة القصيرة**: عند قراءة خلية تتجاوز النهاية المادية لتدفق قابل للنمو،
   تُملأ البايتات الناقصة بالأصفار، مثل سلوك تدفقات Kotlin في الذاكرة.
5. **BinaryList** يستخدم إزاحات 32-بت كما في الصيغة الأصلية، لذا الحد العملي
   لحجم التدفق نحو 2 GiB.
6. **لا انعكاس للتسلسل**: راجع [الانتقال من Kotlin](docs/migration.ar.md).
