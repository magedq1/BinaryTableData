# قوائم البلوبات BinaryList

[English](binarylist.md) | **العربية**

الحزمة: `binarytable`

`BinaryList` تحفظ بيانات متغيرة الطول (صور، مستندات، رسائل...) في تدفق واحد،
مع فهرس منفصل. على عكس `BinaryTableData` لا تحتاج إلى صفوف ثابتة.

---

## 1. الصيغة

```text
offset 0                       int32  realSize
offset 100                     int32  start offset للعنصر 0
...
offset 100 + maxListSize*4     بداية البيانات
```

- `schemaSize = 100` بايت.
- `indexSize = maxListSize * 4` بايت.
- تبدأ البيانات عند `max(streamLength, 100 + indexSize)`.
- نهاية العنصر هي بداية العنصر التالي، أو نهاية التدفق للعنصر الأخير.
- الإزاحات 32-بت، فالحد العملي نحو 2 GiB.

هذه الصيغة مطابقة تماماً لـ `BinaryList.kt`، ويمكن قراءة القوائم بين اللغتين.

### السعة ليست حقلاً مخزّناً

لا تحتوي صيغة Kotlin حقلاً صريحاً لـ `maxListSize`. القائمة الجديدة في هذا
المنفذ تكتب `realSize = 0` وتحجز `100 + maxEntries × 4` بايت. هذا يسمح لـ
`GuessListCapacity` باستعادة السعة من البادئة حتى لو كانت القائمة فارغة، ومن
أول إزاحة بيانات في القوائم القديمة الممتلئة.

```go
list, _ := binarytable.NewBinaryList(stream, 5000)
err := list.Initialize() // حجز البادئة لملف جديد

capacity, _ := binarytable.GuessListCapacity(reader) // 5000
```

---

## 2. الإنشاء

```go
stream := binarytable.NewAutoSizeStream(1024)
list, err := binarytable.NewBinaryList(stream, 1000)
```

الوسيط الثاني هو `maxListSize`. إذا مررت قيمة غير موجبة يُستخدم 1000.
قراءة `realSize` تتم من أول 4 بايت في التدفق، لذا يمكن إعادة فتح قائمة موجودة:

```go
list, err := binarytable.NewBinaryList(
	binarytable.NewByteArrayStreamFixed(existingBytes),
	1000,
)
```

---

## 3. الإضافة والقراءة

```go
index, err := list.Add([]byte("maged"))
index, err = list.AddReader(reader) // بث مباشر دون تحميل البلوب كاملاً
```

`Add` تعيد فهرس العنصر الجديد، أو خطأ إذا امتلأت القائمة. `AddReader` تبث
البيانات مباشرة إلى القائمة، لذا لا تحتاج الملفات الكبيرة إلى مساحة في الذاكرة.

```go
index, err := list.AddFunc(func(w io.Writer) error {
	_, err := w.Write([]byte("ahmed"))
	return err
})
```

`AddFunc` تبني المحتوى في الذاكرة ثم تضيفه. مكافئة لـ
`add { outputStream -> ... }` في Kotlin.

```go
data, err := list.Bytes(index)
reader, err := list.Reader(index) // SubReader
ioReader, err := list.Open(index) // io.Reader تسلسلي
```

- `Bytes` تعيد نسخة كاملة من البلوب.
- `Reader` تعيد `SubReader` للقراءة دون نسخ.
- `ToBytes` تعيد نسخة من التدفق كاملاً.
- `Size` تعيد عدد العناصر.
- `SchemaSize` و`IndexSize` و`MaxSize` متاحة للفحص.

---

## 4. البلوبات الفارغة

إضافة بلوب فارغ لا تكتب بيانات، لذا لا يزيد الطول المادي للتدفق. لمنع مؤشر
غير صالح، تمدد `BinaryList` التدفق إلى بداية البيانات باستخدام واجهة `Grower`.
التدفقات المدمجة (`AutoSizeStream`, `ByteArrayStream`, `FileStream`) تدعم ذلك.

إذا استخدمت تدفقاً خارجياً لا يدعم `Grower`، فإن `Add(nil)` يعيد خطأ واضحاً
بدلاً من كتابة فهرس تالف.

---

## 5. الأخطاء

| الحالة | الخطأ |
| --- | --- |
| الفهرس خارج المدى | `ErrOutOfBounds` |
| القائمة ممتلئة | `ErrOutOfBounds` |
| بداية غير صالحة في الفهرس | `ErrCorrupt` |
| نهاية أصغر من البداية | `ErrCorrupt` |
| تدفق `nil` | خطأ عادي |
| تدفق لا يدعم البلوب الفارغ | خطأ عادي |

---

## 6. مثال مع جدول

تخزين صورة في `BinaryList` وربطها بصف في `BinaryTableData`:

```go
avatarID, err := list.Add(imageBytes)
if err != nil {
	return err
}

if err := table.SetIntByName(row, "avatar_id", avatarID); err != nil {
	return err
}

// لاحقاً
id, err := table.GetIntByName(row, "avatar_id")
image, err := list.Bytes(int(id))
```

هذا النمط يحافظ على صفوف الجدول صغيرة وثابتة، ويضع البيانات الكبيرة في قائمة
منفصلة، تماماً كما في التصميم الأصلي.
