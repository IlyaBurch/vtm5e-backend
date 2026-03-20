# Задание для бэкенд-агента — Cainite.ru

## Контекст

Проект cainite.ru — хранилище чарников для настольных RPG.
Текущий бэкенд уже работает и имеет базовую авторизацию и ручки для VtM.
Задача: прикрутить Google OAuth, добавить поле `username` к пользователю, и сделать отдельные ручки для Mork Borg.

---

## 1. Google OAuth

### Что нужно

Интегрировать Google OAuth 2.0 через Passport.js (если Node.js) или аналог.

### Переменные окружения (добавить в .env)

```env
GOOGLE_CLIENT_ID=<получить в Google Cloud Console>
GOOGLE_CLIENT_SECRET=<получить в Google Cloud Console>
GOOGLE_CALLBACK_URL=https://cainite.ru/auth/google/callback
```

### Ручки

```
GET  /auth/google                  → редирект на Google OAuth
GET  /auth/google/callback         → колбэк после авторизации Google
                                     → при успехе: редирект на /characters
                                     → при ошибке: редирект на /auth?error=google_failed
```

### Логика колбэка

1. Получить профиль из Google (`id`, `email`, `displayName`, `picture`)
2. Поискать пользователя по `googleId` или `email`
3. Если пользователь существует → обновить `googleId` если не был, залогинить
4. Если не существует → создать нового:
   - `email` = из Google
   - `googleId` = из Google
   - `username` = из `displayName` (очистить, сделать уникальным если занят: добавить `_2`, `_3` и т.д.)
   - `avatarUrl` = из Google picture
   - `passwordHash` = null (OAuth-пользователь, пароль не нужен)
5. Вернуть JWT-токен в httpOnly cookie

---

## 2. Поле username у пользователя

### Изменения в модели User

```typescript
// Добавить к существующей модели User:
interface UserUpdate {
  username: string        // уникальный, 3-30 символов, только a-z 0-9 _-
  avatarUrl?: string      // URL аватарки (из Google или загруженная)
  googleId?: string       // ID из Google OAuth
  // passwordHash остаётся nullable — для OAuth-пользователей без пароля
}
```

### Валидация username

- Минимум 3, максимум 30 символов
- Только `a-z`, `0-9`, `_`, `-`
- Уникальный в системе
- Нельзя: `admin`, `root`, `cainite`, `api`, `auth`, `static` (зарезервированные)

### Новые ручки для пользователя

```
GET    /api/user/me                → текущий пользователь
PATCH  /api/user/me                → обновить профиль (username, avatarUrl)

Body PATCH /api/user/me:
{
  "username": "string",      // опционально
  "avatarUrl": "string"      // опционально
}

Response 200:
{
  "id": "uuid",
  "email": "string",
  "username": "string",
  "avatarUrl": "string | null",
  "createdAt": "ISO date"
}

Response 422 (ошибка валидации):
{
  "error": "USERNAME_TAKEN" | "USERNAME_INVALID" | "USERNAME_TOO_SHORT" | "USERNAME_TOO_LONG"
}
```

### Изменения в /auth/register

Добавить поле `username` в тело запроса:

```
POST /auth/register
Body:
{
  "email": "string",
  "username": "string",   // ← новое обязательное поле
  "password": "string"
}
```

---

## 3. Ручки для Mork Borg

Создать отдельный роутер `/api/morkborg/` — не переиспользовать VtM роутер.

### Модель MorkBorgCharacter

```typescript
interface MorkBorgCharacter {
  id:          string     // uuid
  userId:      string     // FK → User
  name:        string
  className:   string     // Класс персонажа
  description: string
  coldBlood:   string     // Хладнокровие (текст, напр. "РС 12 или -д2 ОЗ")
  exhaustion:  string     // Лишение сил (текст, напр. "1 час")

  // Способности класса (JSON массив)
  abilities: { name: string; description: string }[]

  // Характеристики (сырые значения, модификаторы считаются на фронте)
  strength:  number       // Сила      (3-18)
  agility:   number       // Ловкость  (3-18)
  toughness: number       // Стойкость (3-18)
  presence:  number       // Присутствие (3-18)

  // Очки здоровья
  hp:    number
  maxHp: number

  // Оружие (JSON массив, макс 2)
  weapons: { name: string; damage: string; notes: string }[]

  // Броня
  armorName:  string
  armorTier:  string      // "-д2", "-д4", "-д6", "нет"

  // Страдания (JSON массив из 6 булевых)
  sufferings: boolean[]

  // Снаряжение (JSON массив строк, макс 9)
  equipment: string[]

  silver: number

  // Знамения
  omens:     number       // Всего (по классу)
  omensUsed: number       // Использовано

  notes: string

  createdAt: Date
  updatedAt: Date
}
```

### CRUD ручки

```
GET    /api/morkborg/characters
       → список персонажей текущего пользователя
       Response: { characters: MorkBorgCharacter[] }

POST   /api/morkborg/characters
       Body: MorkBorgCharacter (без id, userId, createdAt, updatedAt)
       Response 201: { character: MorkBorgCharacter }

GET    /api/morkborg/characters/:id
       Response 200: { character: MorkBorgCharacter }
       Response 404: { error: "NOT_FOUND" }
       Response 403: { error: "FORBIDDEN" } — если не владелец

PATCH  /api/morkborg/characters/:id
       Body: частичный MorkBorgCharacter (любые поля)
       Response 200: { character: MorkBorgCharacter }

DELETE /api/morkborg/characters/:id
       Response 204: (пусто)
```

### Валидация

```typescript
// Минимальные требования для создания
const MorkBorgCreateSchema = {
  name:      { type: 'string', minLength: 1, maxLength: 100, required: true },
  className: { type: 'string', maxLength: 200 },
  strength:  { type: 'number', min: 1, max: 20 },
  agility:   { type: 'number', min: 1, max: 20 },
  toughness: { type: 'number', min: 1, max: 20 },
  presence:  { type: 'number', min: 1, max: 20 },
  hp:        { type: 'number', min: 0 },
  maxHp:     { type: 'number', min: 1 },
  sufferings:{ type: 'array', items: 'boolean', length: 6 },
  equipment: { type: 'array', items: 'string', maxLength: 9 },
  silver:    { type: 'number', min: 0 },
  omens:     { type: 'number', min: 0, max: 10 },
  omensUsed: { type: 'number', min: 0 },
  weapons:   { type: 'array', maxLength: 2 },
}
```

---

## 4. Миграции БД

Если используется Prisma — добавить в schema.prisma:

```prisma
model User {
  // ... существующие поля ...
  username    String?   @unique
  avatarUrl   String?
  googleId    String?   @unique

  morkBorgCharacters MorkBorgCharacter[]
}

model MorkBorgCharacter {
  id          String   @id @default(uuid())
  userId      String
  user        User     @relation(fields: [userId], references: [id], onDelete: Cascade)

  name        String
  className   String   @default("")
  description String   @default("")
  coldBlood   String   @default("РС 12 или -д2 ОЗ")
  exhaustion  String   @default("1 час")
  abilities   Json     @default("[]")

  strength    Int      @default(10)
  agility     Int      @default(10)
  toughness   Int      @default(10)
  presence    Int      @default(10)

  hp          Int      @default(8)
  maxHp       Int      @default(8)

  weapons     Json     @default("[]")
  armorName   String   @default("")
  armorTier   String   @default("-д2")
  sufferings  Json     @default("[false,false,false,false,false,false]")
  equipment   Json     @default("[]")

  silver      Int      @default(0)
  omens       Int      @default(3)
  omensUsed   Int      @default(0)

  notes       String   @default("")

  createdAt   DateTime @default(now())
  updatedAt   DateTime @updatedAt
}
```

Если используется другой ORM — адаптировать схему аналогично.

---

## 5. Порядок выполнения

1. Добавить поля `username`, `avatarUrl`, `googleId` к модели User + миграция
2. Обновить `/auth/register` — добавить обязательный `username`
3. Создать модель `MorkBorgCharacter` + миграция
4. Реализовать CRUD `/api/morkborg/characters`
5. Настроить Google OAuth + ручки `/auth/google` и `/auth/google/callback`
6. Добавить `/api/user/me` GET и PATCH
7. Протестировать: регистрация с username, Google OAuth, CRUD Mork Borg

---

## 6. Что НЕ трогать

- Существующие VtM ручки (`/api/vtm/...` или как они называются сейчас)
- Существующую логику JWT
- Существующие тесты

Если текущий бэкенд использует другой стек (не Node.js/Prisma) — адаптируй схему и ручки под него, сохранив все имена эндпоинтов и структуру тела запросов/ответов без изменений.
