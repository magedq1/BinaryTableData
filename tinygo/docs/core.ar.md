# الجدول والرأس والأعمدة

[English](core.md) | **العربية**

الحزمة: `binarytable`

---

## 1. تعريف الأعمدة

```go
col := binarytable.IntColumn("_id")          // 4 بايت
col = binarytable.LongColumn("created")      // 8 بايت
col = binarytable.DoubleColumn("score")      // 8 بايت
col = binarytable.StringColumn("name", 64)   // 64 بايت UTF-8
col = binarytable.RawColumn("avatar", 8192)  // 8192 بايت خامة
```

يمكن أيضاً إنشاء عمود مباشرة:

```go
col := binarytable.Column{
	Name:   "custom",
	Length: 32,
	Type:   binarytable.TypeRaw,
}
```

الثوابت المتاحة: `TypeInt`, `TypeLong`, `TypeDouble`, `TypeString`, `TypeRaw`.

### القيود

- اسم العمود لا يمكن أن يكون فارغاً أو يحتوي `/` أو سطراً جديداً.
- الطول يجب أن يكون أكبر من صفر.
- `INT` يحتاج 4 بايت على الأقل، و`LONG`/`DOUBLE` تحتاج 8 بايت على الأقل.
- لا يُسمح بتكرار الأسماء.

---

## 2. إنشاء الرأس

```go
header, err := binarytable.NewHeader(1000, []binarytable.Column{
	binarytable.IntColumn("_id"),
	binarytable.StringColumn("first_name", 64),
})
if err != nil {
	// خطأ في المخطط
}
```

### دوال الرأس

| الدالة | النتيجة |
| --- | --- |
| `MaxRows()` | سعة الصفوف |
| `ColumnCount()` | عدد الأعمدة |
| `Columns()` | نسخة من قائمة الأعمدة |
| `SchemaSize()` | حجم الرأس بالبايت، بما فيه أول 8 بايت |
| `MaxRowLength()` | مجموع أطوال الأعمدة |
| `Capacity()` | حجم الجدول الكامل بالبايت |
| `SchemaBytes()` | نسخة من نص المخطط |
| `ColumnInfo(index)` | معلومات عمود بالفهرس |
| `ColumnInfoByName(name)` | معلومات عمود بالاسم |
| `ColumnIndex(name)` | فهرس العمود |

الرأس غير قابل للتعديل بعد إنشائه. لا تعدّل الشريحة التي تعيدها `Columns()`.

---

## 3. إنشاء الجدول وفتحه

### جدول في الذاكرة

```go
stream := binarytable.NewAutoSizeStream(1024)
table, err := binarytable.New(header, stream)
```

`New` تكتب الرأس تلقائياً، وتستخدم التدفق للقراءة والكتابة.

### جدول بقراءة وكتابة منفصلتين

```go
table, err := binarytable.NewWithStreams(header, reader, writer)
```

### جدول قراءة فقط من تدفق موجود

```go
table, err := binarytable.OpenReader(header, reader)
```

### قراءة الرأس من تدفق

```go
table, err := binarytable.FromReader(reader)
```

### من بايتات

```go
table, err := binarytable.FromBytes(data) // للقراءة فقط
```

### من ملف

```go
table, err := binarytable.CreateFile(header, "users.db") // إنشاء/اقتطاع
defer table.Close()

table, err = binarytable.OpenFile("users.db")           // قراءة وكتابة
defer table.Close()

table, err = binarytable.OpenFileReadOnly("users.db")   // قراءة فقط
defer table.Close()
```

`CreateFile` ينشئ ملفاً بحجم `header.Capacity()` بالضبط. الملفات المفتوحة
بـ `OpenFile` أو `CreateFile` تُغلق عبر `Close`؛ التدفقات التي تملكها أنت
لا تُغلق تلقائياً.

---

## 4. الوصول إلى القيم

### خام

```go
err := table.SetValue(row, column, rawBytes)
err = table.SetValueByName(row, "payload", rawBytes)

raw, err := table.GetValue(row, column)
raw, err = table.GetValueByName(row, "payload")

stream, err := table.GetValueAsStreamByName(row, "payload")
```

`GetValue` يعيد نسخة بطول العمود كامل. `SetValue` يرفض القيم الأطول من العمود
ويعيد `ErrOverSize`.

### الأنواع المكتوبة

| النوع | قراءة | كتابة |
| --- | --- | --- |
| `int32` | `GetInt`, `GetIntByName` | `SetInt`, `SetIntByName` |
| `int64` | `GetLong`, `GetLongByName` | `SetLong`, `SetLongByName` |
| `float64` | `GetDouble`, `GetDoubleByName` | `SetDouble`, `SetDoubleByName` |
| `string` | `GetString`, `GetStringByName` | `SetString`, `SetStringByName` |

مثال:

```go
if err := table.SetInt(0, 0, 42); err != nil { ... }
id, err := table.GetInt(0, 0)

if err := table.SetStringByName(0, "name", "Alice"); err != nil { ... }
name, err := table.GetStringByName(0, "name")
```

`SetString` يملأ بقية العمود بالأصفار، وهو أفضل من سلوك Kotlin عند الكتابة فوق
نص أطول.

---

## 5. المواقع والحجم

```go
pos, err := table.Position(3, 1)
```

يعيد الإزاحة المطلقة للخلية. الصفوف المفهرسة من 0 إلى `MaxRows()-1`، والأعمدة
من 0 إلى `ColumnCount()-1`.

```go
rows, err := table.GetAllRows()
```

يعيد كل البايتات بعد المخطط.

---

## 6. البحث

```go
row, err := table.IndexOfInt(value, column, start)
row, err = table.IndexOfIntByName(value, "id", 0)
row, err = table.IndexOfLong(value, column, start)
row, err = table.IndexOfDouble(value, column, start)
row, err = table.IndexOfString(value, column, start)
```

تعيد `-1` عند عدم العثور على القيمة. `start` هو صف البداية.

مثال:

```go
row, err := table.IndexOfStringByName("Alice", "first_name", 0)
if err != nil {
	return err
}
if row >= 0 {
	fmt.Println("found at row", row)
}
```

---

## 7. الأخطاء

```go
var (
	binarytable.ErrReadOnly
	binarytable.ErrOutOfBounds
	binarytable.ErrOverSize
	binarytable.ErrCorrupt
	binarytable.ErrNoReader
)
```

استخدم `errors.Is`:

```go
if err := table.SetString(row, col, value); errors.Is(err, binarytable.ErrOverSize) {
	// القيمة أطول من العمود
}
```

---

## 8. الربط اليدوي بهيكل Go

لا يوجد تسلسل تلقائي بالانعكاس في TinyGo. اكتب دالتَي تحويل صغيرتين:

```go
type User struct {
	ID    int32
	Name  string
	Score float64
}

func writeUser(table *binarytable.Table, row int, user User) error {
	if err := table.SetIntByName(row, "_id", user.ID); err != nil {
		return err
	}
	if err := table.SetStringByName(row, "name", user.Name); err != nil {
		return err
	}
	return table.SetDoubleByName(row, "score", user.Score)
}

func readUser(table *binarytable.Table, row int) (User, error) {
	id, err := table.GetIntByName(row, "_id")
	if err != nil {
		return User{}, err
	}
	name, err := table.GetStringByName(row, "name")
	if err != nil {
		return User{}, err
	}
	score, err := table.GetDoubleByName(row, "score")
	if err != nil {
		return User{}, err
	}
	return User{ID: id, Name: name, Score: score}, nil
}
```

هذا النمط متوافق مع TinyGo، أسرع من الانعكاس، وواضح عند مراجعة المخطط.
