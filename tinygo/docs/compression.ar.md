# الضغط Deflate

[English](compression.md) | **العربية**

الحزمة: `binarytable`

يوجد مستويان للضغط:

1. **ضغط/فك ضغط مباشر** لشريحة بايتات باستخدام zlib.
2. **تدفق Deflate ذو الكتل** المتوافق مع `DeflateBinaryStreamReader` و
   `DeflateBinaryStreamWriter` في Kotlin.

---

## 1. الضغط المباشر

```go
compressed, err := binarytable.DeflateData(data)
restored, err := binarytable.InflateData(compressed)
```

مع نطاق محدد:

```go
compressed, err := binarytable.DeflateDataRange(data, start, length)
```

- يستخدم `compress/zlib`.
- متوافق مع `java.util.zip.Deflater` (zlib وليس raw deflate).
- `InflateData` يعيد `ErrCorrupt` عند بيانات غير صالحة.

---

## 2. تدفق Deflate ذو الكتل

### الكتابة

```go
source := binarytable.NewByteArrayReader(data)
output := binarytable.NewAutoSizeStream(4096)

err := binarytable.Deflate(
	source,
	output,
	binarytable.DefaultDeflateBlockSize, // 8192
	binarytable.DefaultDeflateHeaderSize, // 0
)
```

`Deflate` تقرأ من `Reader` وتكتب إلى `Writer`، وتترك `headerSize` بايتاً في
البداية لاستخدام المتصل.

### القراءة

```go
reader, err := binarytable.NewDeflateReader(output, 8) // 8 كتل في الكاش
if err != nil {
	return err
}
data, err := binarytable.ReadAll(reader)
```

أو باختصار:

```go
data, err := binarytable.InflateStream(output, 8)
```

- `reader.Size()` تعيد الطول الأصلي قبل الضغط.
- `reader.ReadAt` تعمل عبر حدود الكتل.
- الكتل المفكوكة تُخزّن في LRU cache، ويمكن ضبط عددها.
- تمرير `cacheMax <= 0` يعطّل الكاش.

يمكن قراءة الرأس دون فتح قارئ:

```go
header, err := binarytable.ReadDeflateHeader(source)
fmt.Println(header.BlockSize, header.Blocks, header.OriginalLength)
```

`ReadDeflateHeader` هو المكافئ المباشر لبيانات `DeflateBinaryStreamReader`.
ويعرضه CLI عبر `deflate info` و`deflate read`.

بما أن `DeflateReader` يطبّق `binarytable.Reader`، يمكن تمريره مباشرة إلى
`FromReader` أو تغليفه بـ `AsStream` لـ `BinaryList`. تستخدم الواجهة ذلك
لتصفح الجداول والقوائم المضغوطة للقراءة فقط دون استخراج الملف أولاً.

---

## 3. الصيغة

```text
100  int32  blockSize
104  int32  headerSize
108  int32  maxBlocksCount
112  int32  blocksCount
116  int64  originalLength
headerSize + 512                    فهرس الكتل
headerSize + 512 + maxBlocksCount*16  الكتل المضغوطة
```

كل عنصر فهرس: `int64 start` و`int64 end` نسبيان إلى بداية الكتل.
كل كتلة مضغوطة بـ zlib.

التفاصيل الكاملة في [صيغة البيانات](format.ar.md).

---

## 4. اختيار حجم الكتلة

- القيمة الافتراضية 8192 بايت، وهي نفس قيمة Kotlin.
- كتل أصغر: وصول أسرع لبيانات متفرقة، لكن فهرساً أكبر وضغطاً أقل.
- كتل أكبر: ضغط أفضل، لكن ذاكرة أكبر لكل كتلة مفكوكة.
- عند القراءة العشوائية على متحكم بذاكرة محدودة، استخدم كتلاً صغيرة وكاشاً
  صغيراً (مثلاً 2–4).

---

## 5. مثال

```go
raw := make([]byte, 20000)
for i := range raw {
	raw[i] = byte(i % 64)
}

compressed := binarytable.NewAutoSizeStream(1024)
if err := binarytable.Deflate(
	binarytable.NewByteArrayReader(raw),
	compressed,
	1024,
	0,
); err != nil {
	return err
}

restored, err := binarytable.InflateStream(compressed, 4)
if err != nil {
	return err
}
if !bytes.Equal(raw, restored) {
	return errors.New("compression round trip failed")
}
```

---

## 6. التوافق

- `deflate.bin` في `testdata/` أنشأته Kotlin، ويقرأه Go في
  `TestKotlinDeflateFixture`.
- تم التحقق يدوياً من أن Kotlin يقرأ تدفقاً كتبه Go.
- محتوى zlib نفسه قد يختلف قليلاً بين Java وGo من حيث البايتات، لكن الصيغة
  القابلة للقراءة متوافقة. المهم أن الطول الأصلي وفهرس الكتل صحيحان.
