# TinyGo: القيود والاستهداف

[English](tinygo.md) | **العربية**

هذه الصفحة تشرح كيف صُمم المنفذ ليعمل مع TinyGo، وما الذي يجب مراعاته عند
الانتقال من Go القياسي إلى TinyGo.

---

## 1. ما الذي يعمل؟

| الميزة | TinyGo |
| --- | --- |
| تدفقات الذاكرة | ✅ |
| `ByteArrayReader` / `AutoSizeStream` | ✅ |
| جدول ثابت المخطط | ✅ |
| `BinaryList` | ✅ |
| ضغط zlib / Deflate ذو الكتل | ✅ |
| ملفات عبر `os.File` | ✅ على الأهداف المستضافة |
| AES / SHA-256 / عشوائية | ✅ |
| RSA / X.509 | ⚠️ ثقيل، اختبره على هدفك |
| Serialization بالانعكاس | ❌ |
| mmap | ❌ (مستبدل بـ ReadAt/WriteAt) |

الهدف الأساسي هو **ألا يعتمد جوهر المكتبة على `reflect` أو `unsafe`**.
والحزم القياسية المستخدمة موجودة في قائمة TinyGo المدعومة:

`bytes`, `compress/zlib`, `container/list`, `crypto/aes`, `crypto/cipher`,
`crypto/rand`, `crypto/rsa`, `crypto/sha256`, `crypto/x509`, `encoding/pem`,
`errors`, `fmt`, `io`, `math`, `os`, `strconv`, `strings`, `sync`, `testing`.

**واجهة Wails تطبيق سطح مكتب** وليست هدف TinyGo. تعتمد على WebView2 وWails،
بينما تبقى المكتبة الأساسية متوافقة مع TinyGo.

---

## 2. لا Serialization تلقائي

في Kotlin:

```kotlin
@Serializable
data class User(val id: Int, @BinaryColumnLength(64) val name: String)

val table = BinaryTableData.create<User>(file, 1000)
table.encodeRow(0, User(1, "Alice"))
```

في Go/TinyGo لا يوجد مكافئ آمن للانعكاس. البديل:

```go
header, _ := binarytable.NewHeader(1000, []binarytable.Column{
	binarytable.IntColumn("id"),
	binarytable.StringColumn("name", 64),
})

func writeUser(t *binarytable.Table, row int, u User) error {
	if err := t.SetIntByName(row, "id", u.ID); err != nil {
		return err
	}
	return t.SetStringByName(row, "name", u.Name)
}
```

هذا يعني كتابة كود ربط صغير، لكنه:
- يعمل على TinyGo.
- أسرع من الانعكاس.
- واضح عند تغيير المخطط.
- لا يضيف تبعية `kotlinx.serialization`.

يمكن توليد دوال الربط تلقائياً بأداة codegen خاصة بك إذا كان المشروع كبيراً.

---

## 3. الذاكرة والكتل

- `AutoSizeStream` يخصص كتلاً بحجم `blockSize` (الافتراضي 1024).
- في متحكم بذاكرة صغيرة، استخدم كتلاً أصغر وحدد سعة الجدول بدقة.
- `ByteArrayReader` لا ينسخ البيانات؛ استخدمه لقراءة ثابتة.
- `GetValueAsStream` يتجنب نسخ الأعمدة الخام الكبيرة.
- `DeflateReader` يحتفظ بعدد محدود من الكتل المفكوكة. اضبط `cacheMax` صغيراً
  (2–4) على الأجهزة المحدودة.
- `EncryptStream` يقرأ كل المدخل في الذاكرة. للبيانات الضخمة استخدم `Encrypt`
  على قطع.

---

## 4. الملفات والأهداف

`OpenFile`, `CreateFile`, `OpenFileStream`, `CreateFileStream` تعتمد على
`os.File`. تعمل على:

- Linux
- Windows
- macOS
- WASI

ولا تُستخدم عادةً على أهداف bare-metal. استخدم تدفقات الذاكرة هناك، أو نفّذ
`Reader`/`Writer` فوق وسيط التخزين الخاص بك (Flash، SD، EEPROM...). الواجهات
صغيرة:

```go
type Reader interface {
	ReadAt(p []byte, off int64) (int, error)
	Size() int64
}

type Writer interface {
	WriteAt(p []byte, off int64) (int, error)
}
```

أي وسيط يطبّق هاتين الدالتين يمكن استخدامه مع الجدول و`BinaryList` والضغط.

---

## 5. أوامر البناء

من داخل `tinygo/`:

```bash
# Go القياسي
go build ./...
go test ./...

# TinyGo لهدف WASI
tinygo build -target=wasi -o demo.wasm ./examples/demo

# TinyGo لهدف Linux (إن كان متاحاً)
tinygo build -target=linux -o demo ./examples/demo

# اختبارات TinyGo حسب الهدف
tinygo test ./...
```

### تقليل الحجم

- لا تستورد `cryptoutil` إن لم تحتج التشفير.
- استخدم `-opt=z` أو `-opt=s` حسب الهدف.
- تجنّب `fmt` في المسار الساخن على المتحكمات.
- استخدم `-print-allocs=.` لفحص التخصيصات (حسب إصدار TinyGo).

---

## 6. ما الذي قد يختلف؟

- `fmt` و`reflect` في TinyGo قد لا يدعمان كل الحالات المتقدمة. لا تعتمد على
  `reflect` في منطق المكتبة.
- بعض الأهداف لا تدعم `recover` بالكامل. المكتبة لا تعتمد على `recover`.
- `os` قد يكون محدوداً على بعض الأهداف.
- `crypto/rsa` قد يعمل على أهداف مستضافة فقط. اختبره.
- أداء الخرائط (`map`) أبطأ نسبياً؛ أسماء الأعمدة تُحل مرة واحدة عند بناء
  الرأس، ثم تُستخدم قائمة `ColumnInfo` للفهرسة.

---

## 7. توصية عملية

إذا كان هدفك متحكماً بذاكرة محدودة:

1. استخدم مخططاً ثابتاً صغيراً.
2. استخدم `AutoSizeStream` أو تدفق Flash خاص بك.
3. استخدم `BinaryList` للبيانات الكبيرة بدل أعمدة RAW ضخمة.
4. استخدم ضغط Deflate فقط إذا كانت البيانات قابلة للضغط.
5. لا تستخدم RSA على الجهاز؛ استخدم AES أو تحقق من توقيع مسبق.
