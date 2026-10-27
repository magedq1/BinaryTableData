# صيغة البيانات والتوافق

[English](format.md) | **العربية**

كل الأعداد في هذه المكتبة **Big-endian**، تماماً كما في الإصدار Kotlin الأصلي.
هذا القرار هو أساس التوافق الثنائي بين اللغتين.

---

## 1. رأس الجدول (Table Header)

الجدول يبدأ بالبنية التالية:

| الإزاحة | الحجم | النوع | المعنى |
| --- | --- | --- | --- |
| `0` | 4 | `int32` | `maxRowsCount` |
| `4` | 4 | `int32` | حجم نص المخطط بالبايت |
| `8` | متغير | UTF-8 | نص المخطط |

نص المخطط سطر لكل عمود:

```
name/length/TYPE\n
```

مثال:

```
_id/4/INT
first_name/16/STRING
last_name/16/STRING
created/8/LONG
score/8/DOUBLE
payload/8/RAW
```

القيم المقبولة لـ `TYPE` هي: `INT`, `LONG`, `DOUBLE`, `STRING`, `RAW`.
هذه هي نفس القيم الناتجة عن `BinaryColumn.ColumnType.name` في Kotlin.

### حساب الإزاحات

```text
schemaSize  = 8 + len(schema)
rowLength   = مجموع أطوال كل الأعمدة
cellOffset  = schemaSize + rowIndex*rowLength + columnStartPosition
capacity    = schemaSize + maxRowsCount*rowLength
```

`columnStartPosition` هو مجموع أطوال الأعمدة السابقة.

---

## 2. أنواع الأعمدة

| النوع | الطول الافتراضي | الترميز |
| --- | --- | --- |
| `INT` | 4 بايت | `int32` بترتيب Big-endian |
| `LONG` | 8 بايت | `int64` بترتيب Big-endian |
| `DOUBLE` | 8 بايت | IEEE-754 Big-endian |
| `STRING` | تحدده أنت | UTF-8 داخل حقل ثابت الطول |
| `RAW` | تحدده أنت | بايتات خامة |

### النصوص

- `SetString` يكتب النص ثم يملأ بقية العمود بالأصفار.
- `GetString` يحذف كل بايتات `0x00` من النتيجة، مثل
  `getString` في Kotlin.
- الطول يُحسب بالبايتات وليس بعدد الحروف. النص العربي قد يحتاج ضعف عدد
  الحروف تقريباً، لذا اختر طولاً كافياً.

### الخلايا غير المكتوبة

الملف أو التدفق يُنشأ بأصفار. أي خلية لم تُكتب بعد تكون قيمتها صفراً، وهذا
مطابق لسلوك Kotlin.

---

## 3. الجداول القابلة للنمو والجداول الثابتة

هناك طريقتان لإنشاء الجدول:

1. **تدفق ذاكرة قابل للنمو** مثل `AutoSizeStream`: الطول الفعلي ينتهي عند آخر
   بايت كُتب، لكن القراءة تتجاوز النهاية وتعيد أصفاراً.
2. **ملف ثابت السعة** عبر `CreateFile`: يُنشأ الملف بحجم
   `header.Capacity()` بالضبط، تماماً مثل
   `BinaryStreamImplMapped(file, schemaSize + maxRows*rowLength)` في Kotlin.

`GetAllRows()` يعيد البيانات من نهاية المخطط حتى نهاية التدفق المادي. في ملف
ثابت السعة يعيد كل صفوف السعة، بينما في تدفق قابل للنمو يعيد فقط ما كُتب.

---

## 4. قائمة البلوبات (BinaryList)

```text
offset 0                         int32  realSize
offset 100                       int32  start offset للعنصر 0
offset 100 + 4                   int32  start offset للعنصر 1
...
offset 100 + maxListSize*4       بداية منطقة البيانات
```

- `schemaSize = 100`.
- `indexSize = maxListSize * 4`.
- تبدأ البيانات عند
  `max(streamLength, 100 + indexSize)`.
- نهاية العنصر هي بداية العنصر التالي، أو نهاية التدفق للعنصر الأخير.
- الإزاحات 32-بت، لذلك لا يمكن أن يتجاوز التدفق نحو 2 GiB.
- عند إضافة بلوب فارغ، يجب تمديد التدفق إلى بداية منطقة البيانات. منفذ Go
  يستخدم واجهة `Grower` تلقائياً مع التدفقات المدمجة.
- `maxListSize` ليس حقلاً مخزناً. القائمة الجديدة التي ينشئها Go تحجز
  `100 + maxListSize*4` بايت، ليستعيدها `GuessListCapacity` لاحقاً؛ وفي القوائم
  القديمة الممتلئة تُستخدم أول إزاحة بيانات.

---

## 5. صيغة Deflate ذات الكتل

منطقة الرأس المحجوزة 512 بايت، ثم فهرس الكتل، ثم الكتل المضغوطة:

| الإزاحة | الحجم | المعنى |
| --- | --- | --- |
| `100` | 4 | `blockSize` |
| `104` | 4 | `headerSize` (بيانات يستخدمها المتصل) |
| `108` | 4 | `maxBlocksCount` |
| `112` | 4 | `blocksCount` |
| `116` | 8 | الطول الأصلي قبل الضغط |
| `headerSize + 512` | `maxBlocksCount*16` | فهرس الكتل |
| `headerSize + 512 + maxBlocksCount*16` | متغير | الكتل |

كل عنصر فهرس يحتوي:

- `int64 start` نسبةً إلى بداية منطقة الكتل.
- `int64 end` نسبةً إلى بداية منطقة الكتل.

كل كتلة مضغوطة باستخدام **zlib** (وليس raw deflate)، لأن
`java.util.zip.Deflater` في Kotlin ينتج zlib افتراضياً. منفذ Go يستخدم
`compress/zlib`، والنتيجة قابلة للقراءة في الاتجاهين.

بايتات `[0, headerSize)` تبقى كما هي، ما يسمح للمتصل بحجز بيانات إضافية.

---

## 6. التحقق من التوافق

ملفات `testdata/` الثنائية تم إنشاؤها بواسطة الكود Kotlin الأصلي:

- `table.bin` — جدول في الذاكرة.
- `table-file.bin` — جدول ملفات ثابت السعة.
- `list.bin` — `BinaryList`.
- `deflate.bin` + `deflate-raw.bin` — تدفق Deflate والبيانات الأصلية.
- `aes.bin` + `aes-iv.bin` + `aes-plain.bin` — AES/CBC/PKCS#5.

اختبارات `compat_test.go` تقرأ هذه الملفات. كذلك تم التحقق يدوياً من الاتجاه
المعاكس: ملفات كتبها Go قرأتها مكتبة Kotlin بنجاح. السكربتات موجودة في
`testdata/kotlin/`.
