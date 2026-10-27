# التشفير AES / RSA

[English](security.md) | **العربية**

الحزمة: `cryptoutil`

هذه الحزمة منفصلة عن `binarytable`، لذا لا تدفع تكلفة التشفير إذا لم تستخدمه.

---

## 1. AES المتوافق مع Kotlin

الإصدار الأصلي يستخدم:

- مفتاحاً مشتقاً من كلمة المرور عبر SHA-256 (32 بايت).
- `AES/CBC/PKCS5Padding`.
- IV بطول 16 بايت.

منفذ Go يطبّق نفس الخطوات، لذا الناتج **مطابق بايتاً ببايت** عند نفس المفتاح
وIV والبيانات. هذا مؤكد باختبار `TestKotlinAESFixture`.

### التشفير

```go
import "github.com/magedq1/BinaryTableData/tinygo/cryptoutil"

iv, err := cryptoutil.GenerateIV()
if err != nil {
	return err
}

ciphertext, err := cryptoutil.Encrypt("my password", iv, plaintext)
```

### فك التشفير

```go
plaintext, err := cryptoutil.Decrypt("my password", iv, ciphertext)
```

### التدفقات

```go
err := cryptoutil.EncryptStream(password, iv, srcReader, dstWriter)
err = cryptoutil.DecryptStream(password, iv, srcReader, dstWriter)
```

`EncryptStream`/`DecryptStream` تقرأ كامل المدخل في الذاكرة لأن CBC يحتاج كتلاً
كاملة. للبيانات الضخمة على هدف مستضاف، استخدم `Encrypt`/`Decrypt` على قطع
مناسبة أو نفّذ حلقة خاصة بك.

### تخزين IV

IV ليس سراً، لكن يجب أن يكون عشوائياً وفريداً لكل عملية تشفير. خزّنه مع
النص المشفر (مثلاً أول 16 بايت من الملف) ولا تعِد استخدامه مع نفس المفتاح.

### اشتقاق المفتاح

```go
key := cryptoutil.PasswordToKey("my password") // 32 بايت
```

هذه دالة SHA-256 بسيطة، وليست KDF مكلفة مثل PBKDF2/Argon2. للأنظمة الحقيقية
التي تحمي كلمات مرور بشرية، استخدم KDF حديثاً. الدالة هنا موجودة للتوافق مع
Kotlin.

---

## 2. RSA

`cryptoutil` يوفّر:

```go
privateKey, err := cryptoutil.GenerateRSAKeyPair(2048)
publicKey := &privateKey.PublicKey

ciphertext, err := cryptoutil.EncryptRSA(publicKey, data)
data, err = cryptoutil.DecryptRSA(privateKey, ciphertext)

signature, err := cryptoutil.SignSHA256(privateKey, data)
err = cryptoutil.VerifySHA256(publicKey, data, signature)
```

تشفير/فك تشفير المفاتيح:

```go
publicPEM, err := cryptoutil.EncodeRSAPublicKeyPEM(publicKey)
privatePEM, err := cryptoutil.EncodeRSAPrivateKeyPEM(privateKey)

publicKey, err = cryptoutil.ParseRSAPublicKeyPEM(publicPEM)
privateKey, err = cryptoutil.ParseRSAPrivateKeyPEM(privatePEM)
```

- التشفير يستخدم PKCS#1 v1.5، ليتوافق مع `Cipher.getInstance("RSA")` في Java.
- التوقيع يستخدم `SHA256withRSA`.
- البيانات المشفرة يجب ألا تتجاوز `publicKey.Size()-11` بايت. للبيانات الأكبر
  استخدم تشفيراً هجيناً: مفتاح AES عشوائي + RSA للمفتاح.
- توليد المفاتيح وX.509 مكلفان على المتحكمات. يُفضّل توفير مفتاح عام مُضمّن
  وعدم توليد مفاتيح على الجهاز.

---

## 3. توليد سلاسل عشوائية

```go
s, err := cryptoutil.RandomString(32)
```

- يستخدم `crypto/rand`.
- نفس أبجدية `EncryptionHelper.randomString` في Kotlin.
- يستخدم rejection sampling لتجنّب انحياز modulo.

---

## 4. قيود TinyGo

- `crypto/aes`, `crypto/cipher`, `crypto/sha256`, `crypto/rand` مدعومة.
- `crypto/rsa` و`crypto/x509` قابلة للاستيراد، لكنها ثقيلة وقد تعتمد على أجزاء
  من `reflect` غير مكتملة في بعض الأهداف. اختبرها على هدفك قبل الاعتماد عليها.
- على متحكم صغير، فضّل AES أو توقيعاً مُتحققاً مسبقاً بدل RSA على الجهاز.
