# الأدوات المساعدة

[English](helpers.md) | **العربية**

الحزمة: `binarytable`

---

## 1. ترميز الأعداد

كل الأعداد Big-endian، كما في Kotlin.

```go
b := binarytable.EncodeInt32(42)
v, err := binarytable.DecodeInt32(b)

b = binarytable.EncodeInt64(1_700_000_000_000)
v64, err := binarytable.DecodeInt64(b)

b = binarytable.EncodeFloat64(12.5)
f, err := binarytable.DecodeFloat64(b)
```

مع إلحاق بدون تخصيص إضافي:

```go
dst = binarytable.AppendInt32(dst, value)
dst = binarytable.AppendInt64(dst, value)
dst = binarytable.AppendFloat64(dst, value)
```

---

## 2. البحث

```go
row, err := table.IndexOfInt(value, column, start)
row, err = table.IndexOfLong(value, column, start)
row, err = table.IndexOfDouble(value, column, start)
row, err = table.IndexOfString(value, column, start)

row, err = table.IndexOfIntByName(value, "id", start)
row, err = table.IndexOfLongByName(value, "created", start)
row, err = table.IndexOfDoubleByName(value, "score", start)
row, err = table.IndexOfStringByName(value, "name", start)
```

- `start` هو أول صف يتم فحصه.
- القيمة `-1` تعني عدم وجود تطابق.
- البحث خطي، مثل `ContentHelper.kt` في Kotlin.

---

## 3. LRU Cache

```go
cache := binarytable.NewCache[string, int](50)
cache.Put("a", 1)

if value, ok := cache.Get("a"); ok {
	fmt.Println(value)
}

cache.Remove("a")
cache.Clear()
```

- متوافق مع `CacheManager` في Kotlin.
- thread-safe.
- السعة `<= 0` تعطل التخزين.
- يُستخدم داخلياً في `DeflateReader`.

---

## 4. أدوات التدفقات

```go
data, err := binarytable.ReadAll(reader)
data, err = binarytable.ReadBytesAt(reader, offset, size)
err = binarytable.ReadAtFull(reader, offset, buffer)
err = binarytable.WriteAtFull(writer, offset, buffer)
err = binarytable.CopyAt(dst, dstOffset, src, srcOffset, size)
```

- `ReadAtFull` يضمن ملء المخزن أو إرجاع خطأ.
- `WriteAtFull` يضمن كتابة كل البايتات.
- `CopyAt` ينسخ على شكل قطع دون تحميل كل البيانات.

---

## 5. الأخطاء المشتركة

```go
binarytable.ErrReadOnly
binarytable.ErrOutOfBounds
binarytable.ErrOverSize
binarytable.ErrCorrupt
binarytable.ErrNoReader
```

```go
if errors.Is(err, binarytable.ErrOverSize) {
	// القيمة أطول من العمود
}
```

---

## 6. التزامن

- `Cache` آمن للاستخدام المتزامن.
- `ByteArrayStream`, `AutoSizeStream`, `FileStream` آمنة.
- `Table` لا يحمل قفلاً داخلياً.
- `DeflateReader` يستخدم `Cache` الآمن، لكن لا تضمن قراءة متزامنة متسقة إذا
  كان المصدر نفسه يتغير أثناء القراءة.

---

## 7. نصائح للأداء

- استخدم `GetValueAsStream` للبيانات الخام الكبيرة بدل نسخها كاملة.
- استخدم `AutoSizeStream` بكتلة مناسبة (1024–8192) لتقليل إعادة التخصيص.
- استخدم `JoinedReader` لقراءة ملفات مقسمة دون دمجها على القرص.
- استخدم `BinaryList.Reader` لتمرير البلوب إلى محلل آخر دون نسخ.
- في TinyGo، تجنّب التخصيصات داخل الحلقات الساخنة قدر الإمكان.
