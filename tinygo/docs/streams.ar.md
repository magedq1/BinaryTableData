# التدفقات والقراءة العشوائية

[English](streams.md) | **العربية**

الحزمة: `binarytable`

---

## 1. الواجهات

```go
type Reader interface {
	ReadAt(p []byte, off int64) (n int, err error)
	Size() int64
}

type Writer interface {
	WriteAt(p []byte, off int64) (n int, err error)
}

type Stream interface {
	Reader
	Writer
}

type Grower interface {
	Grow(size int64) error
}
```

- `ReadAt` تتبع عقد `io.ReaderAt`: تعيد `io.EOF` إذا كانت البايتات المتاحة أقل
  من المطلوب.
- `Size` تعيد الطول المنطقي الحالي.
- `Grower` تستخدمها `BinaryList` لتمديد التدفق عند إضافة بلوب فارغ.

### تحويل قارئ إلى تدفق

```go
stream := binarytable.AsStream(reader)
```

الكتابة على الناتج تعيد `ErrReadOnly`.

---

## 2. تدفقات الذاكرة

### ByteArrayReader

قراءة فقط فوق شريحة موجودة، دون نسخ:

```go
reader := binarytable.NewByteArrayReader(data)
```

### ByteArrayStream

تدفق فوق شريحة ثابتة:

```go
stream := binarytable.NewByteArrayStream(make([]byte, 4096))
```

- `Size` = أعلى موضع كُتب إليه.
- الطول لا يتجاوز حجم الشريحة.
- الكتابة التي تتجاوز النهاية تُقصّ وتعيد `io.ErrShortWrite`.
- الكتابة بعد النهاية المطلقة تعيد `ErrOutOfBounds`.
- `NewByteArrayStreamFixed(data)` يجعل `Size` مساوياً لطول الشريحة دائماً.
- `NewByteArrayStreamCapacity(n)` ينشئ شريحة صفرية بالحجم المطلوب.
- `ToBytes()` يعيد نسخة من البيانات المكتوبة.

### AutoSizeStream

يدعم النمو ويستخدم كتل ثابتة لتقليل إعادة التخصيص:

```go
stream := binarytable.NewAutoSizeStream(1024)
```

- حجم الكتلة الافتراضي 1024 عند تمرير قيمة غير موجبة.
- `ToBytes()` يعيد نسخة من الطول المنطقي.
- `Grow(n)` تمدد الطول بأصفار دون كتابة بيانات.

هذا هو الخيار المناسب للجداول في الذاكرة وللمتحكمات.

---

## 3. تدفق الملفات

```go
stream, err := binarytable.OpenFileStream("data.bin", binarytable.ReadWrite)
defer stream.Close()

stream, err = binarytable.OpenFileStream("data.bin", binarytable.ReadOnly)

stream, err = binarytable.CreateFileStream("data.bin", size, binarytable.ReadWrite)
```

- `CreateFileStream` ينشئ أو يقتطع الملف إلى الحجم المطلوب.
- `binarytable.CreateFile(header, path)` ينشئ ملف جدول بالسعة المناسبة.
- `Grow` تمدد الملف باستخدام `Truncate`.
- `Sync` تفرغ البيانات إلى التخزين.
- `Name` تعيد مسار الملف.

بدلاً من `MappedByteBuffer` في Kotlin، يستخدم Go `ReadAt/WriteAt`. هذا يحافظ
على القراءة العشوائية دون الحاجة إلى mmap، ويعمل على TinyGo المستضاف.

---

## 4. النوافذ والدمج

### SubReader

```go
sub, err := binarytable.NewSubReader(reader, start, end)
sub, err = binarytable.NewSubReader(reader, start) // حتى النهاية
```

- نافذة قراءة فقط.
- `Start` و`End` و`Source` متاحة للفحص.
- `SubReader(start, end)` تنشئ نافذة نسبية.

### JoinedReader

```go
joined := binarytable.NewJoinedReader(part1, part2, part3)
```

تدمج عدة قراء في تدفق منطقي واحد. مفيدة للملفات المقسمة.
القراءة تعبر حدود الأجزاء تلقائياً.

---

## 5. أدوات مساعدة

```go
b, err := binarytable.ReadBytesAt(reader, offset, size)
b, err := binarytable.ReadAll(reader)
err = binarytable.ReadAtFull(reader, offset, buffer)
err = binarytable.WriteAtFull(writer, offset, buffer)
err = binarytable.CopyAt(dst, dstOffset, src, srcOffset, size)

// مكافئات asInputStream / asOutputStream
ioReader := binarytable.NewIOReader(reader)
ioWriter := binarytable.NewIOWriter(writer, startOffset)
```

- `ReadAtFull` و`WriteAtFull` يعيدان خطأ إذا لم يكتمل الطول المطلوب.
- `CopyAt` ينسخ بين تدفقين دون تحميل كل البيانات في الذاكرة.

---

## 6. التزامن

- `ByteArrayStream`, `AutoSizeStream`, `FileStream` آمنة للاستخدام المتزامن.
- `ByteArrayReader` للقراءة فقط ولا تحتاج قفلاً.
- `SubReader` و`JoinedReader` يعتمدان على أمان المصدر.
- `Table` نفسه لا يحمل قفلاً؛ استخدم قفلاً خارجياً عند الكتابة المتزامنة على
  نفس الصفوف.

---

## 7. مثال كامل

```go
src := binarytable.NewAutoSizeStream(256)
_, _ = src.WriteAt([]byte("hello world"), 0)

sub, _ := binarytable.NewSubReader(src, 6, 11)
b, _ := binarytable.ReadAll(sub)
fmt.Println(string(b)) // world

joined := binarytable.NewJoinedReader(
	binarytable.NewByteArrayReader([]byte("abc")),
	binarytable.NewByteArrayReader([]byte("def")),
)
all, _ := binarytable.ReadAll(joined)
fmt.Println(string(all)) // abcdef
```
