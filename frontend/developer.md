# TradeOff Frontend — Developer Guide

## 1. Proje Kurulumu

```bash
cd c:\Users\mehme\brokerApp\frontend
npx -y create-next-app@latest ./ --typescript --tailwind --eslint --app --src-dir --import-alias "@/*" --use-npm
npm install zustand axios lightweight-charts lucide-react framer-motion react-hook-form zod @hookform/resolvers sonner clsx tailwind-merge
```

### Fontlar — `src/app/layout.tsx`
Google Fonts: **Inter** (UI) + **JetBrains Mono** (fiyatlar). `next/font/google` ile yükle.

### Env — `.env.local`
```
NEXT_PUBLIC_API_URL=http://localhost:3000
NEXT_PUBLIC_WS_URL=wss://stream.binance.com:443/ws/btcusdt@trade
```

---

## 2. Tasarım Sistemi

### Renk Paleti (Tailwind `tailwind.config.ts` extend)

```js
colors: {
  brand: { 50:'#EFF6FF', 100:'#DBEAFE', 500:'#3B82F6', 600:'#2563EB', 700:'#1D4ED8' },
  surface: { light:'#FFFFFF', dark:'#111827', card:'#1E2640' },
  success: '#10B981',
  danger: '#EF4444',
  warning: '#F59E0B',
}
```

### Dark Mode
`tailwind.config.ts` → `darkMode: 'class'`. `<html>` elementine `dark` class toggle. Zustand `themeStore` ile yönet, `localStorage` ile persist et.

### Tipografi
- Başlıklar: Inter Bold/Semibold
- Body: Inter Regular 16px
- Fiyatlar: JetBrains Mono Medium
- Caption: Inter Medium 12px

### Spacing: 4px grid. Border-radius: sm=6, md=8, lg=12, xl=16px.

---

## 3. Backend API Kontratları

**Base URL:** `http://localhost:3000`

### Public Endpoints (JWT gerektirmez)

**POST `/api/auth/register`**
```json
// Request
{ "email": "string", "password": "string (min 8)", "first_name": "string", "last_name": "string" }
// Response 201
{ "token": "jwt_string", "user": { "id": "uuid", "email": "str", "first_name": "str", "last_name": "str" } }
```

**POST `/api/auth/login`**
```json
// Request
{ "email": "string", "password": "string" }
// Response 200
{ "token": "jwt_string", "user": { "id": "uuid", "email": "str", "first_name": "str", "last_name": "str" } }
```

### Protected Endpoints (Header: `Authorization: Bearer <token>`)

**GET `/api/wallet`** → `{ id, user_id, balance: "decimal_str", btc_balance: "decimal_str", updated_at }`

**POST `/api/wallet/deposit`** → Body: `{ "amount": "1500.50" }` → Updated wallet

**POST `/api/wallet/withdraw`** → Body: `{ "amount": "500.00" }` → Updated wallet

**POST `/api/orders`**
```json
// Request
{ "symbol": "BTCUSDT", "side": "BUY"|"SELL", "type": "MARKET"|"LIMIT", "quantity": "0.05", "price": "65000.00" }
// Response 201 → Order { id, user_id, symbol, side, type, quantity, price, status:"PENDING", created_at, updated_at }
```

### Error Format
```json
{ "error": "mesaj", "code": "error_code" }
```
Kodlar: `user_not_found(404)`, `user_already_exists(409)`, `invalid_credentials(401)`, `insufficient_balance(422)`, `invalid_amount(400)`

---

## 4. TypeScript Tipleri — `src/types/domain.ts`

```typescript
export interface User {
  id: string; email: string; first_name: string; last_name: string;
}
export interface AuthResponse {
  token: string; user: User;
}
export interface Wallet {
  id: string; user_id: string; balance: string; btc_balance: string; updated_at: string;
}
export type OrderSide = 'BUY' | 'SELL';
export type OrderType = 'MARKET' | 'LIMIT';
export type OrderStatus = 'PENDING' | 'COMPLETED' | 'FAILED' | 'CANCELED';
export interface Order {
  id: string; user_id: string; symbol: string; side: OrderSide;
  type: OrderType; quantity: string; price: string; status: OrderStatus;
  created_at: string; updated_at: string;
}
export interface PlaceOrderRequest {
  symbol: string; side: OrderSide; type: OrderType; quantity: string; price: string;
}
export interface LivePrice {
  symbol: string; price: string; timestamp: number;
}
export interface Transaction {
  id: string; user_id: string; type: 'DEPOSIT'|'WITHDRAWAL'; amount: string;
  currency: string; status: 'PENDING'|'COMPLETED'|'FAILED'; stripe_ref: string; created_at: string;
}
```

---

## 5. Dizin Yapısı

```
src/
├── app/
│   ├── layout.tsx              # Root: providers, fonts, ThemeProvider
│   ├── page.tsx                # Landing page
│   ├── globals.css             # CSS custom properties + base styles
│   ├── not-found.tsx           # 404 sayfası
│   ├── (auth)/
│   │   ├── layout.tsx          # Split-screen auth layout
│   │   ├── login/page.tsx
│   │   └── register/page.tsx
│   └── (dashboard)/
│       ├── layout.tsx          # Sidebar + Topbar layout
│       ├── dashboard/page.tsx
│       ├── trade/page.tsx
│       ├── wallet/page.tsx
│       └── settings/page.tsx
├── components/
│   ├── ui/                     # Button, Input, Card, Modal, Badge, Spinner, Toast, Table, Tabs, Toggle, Skeleton, ThemeToggle
│   ├── layout/                 # Navbar, Sidebar, Topbar, Footer, MobileMenu
│   ├── landing/                # HeroSection, FeatureCard, StatsSection, CTASection
│   ├── auth/                   # LoginForm, RegisterForm
│   ├── dashboard/              # PortfolioCard, PriceCard, QuickTradeWidget, RecentOrdersTable, MiniChart
│   ├── trade/                  # PriceChart, OrderForm, TickerBar, OrderHistory
│   └── wallet/                 # BalanceCards, DepositForm, WithdrawForm, TransactionHistory
├── lib/
│   ├── api/                    # client.ts, auth.ts, wallet.ts, orders.ts
│   ├── ws/binance.ts           # WebSocket manager
│   ├── utils/                  # formatters.ts, validators.ts, cn.ts
│   └── constants.ts
├── hooks/                      # useAuth, useWallet, useLivePrice, useOrders, useTheme
├── stores/                     # authStore, walletStore, tradeStore, themeStore (Zustand)
├── types/                      # domain.ts, api.ts, common.ts
└── middleware.ts                # Route protection
```

---

## 6. Katman Katman Uygulama Sırası

### FAZE 1: Altyapı
1. Next.js projesi oluştur (komut yukarıda)
2. `tailwind.config.ts` — renk paleti, fontlar, dark mode ayarı
3. `globals.css` — CSS custom properties, base reset
4. `src/lib/utils/cn.ts` → `clsx` + `tailwind-merge` wrapper
5. `src/types/domain.ts` — tüm TypeScript tipleri
6. `src/lib/constants.ts` — API_URL, WS_URL

### FAZE 2: API Katmanı
7. `src/lib/api/client.ts` — Axios instance:
   - baseURL: `process.env.NEXT_PUBLIC_API_URL`
   - Request interceptor: localStorage'dan token oku, `Authorization` header ekle
   - Response interceptor: 401 → `authStore.logout()` + `router.push('/login')`
8. `src/lib/api/auth.ts` — `register(data)`, `login(data)` fonksiyonları
9. `src/lib/api/wallet.ts` — `getWallet()`, `deposit(amount)`, `withdraw(amount)`
10. `src/lib/api/orders.ts` — `placeOrder(data)`

### FAZE 3: State Management (Zustand)
11. `src/stores/authStore.ts`:
    ```typescript
    interface AuthState {
      user: User | null; token: string | null; isAuthenticated: boolean;
      login: (res: AuthResponse) => void;
      logout: () => void;
      hydrate: () => void; // localStorage'dan oku
    }
    ```
12. `src/stores/walletStore.ts`:
    ```typescript
    interface WalletState {
      wallet: Wallet | null; loading: boolean;
      fetchWallet: () => Promise<void>;
      setWallet: (w: Wallet) => void;
    }
    ```
13. `src/stores/tradeStore.ts`:
    ```typescript
    interface TradeState {
      livePrice: string; priceHistory: {time:number,value:number}[];
      orders: Order[];
      setLivePrice: (price: string) => void;
      addPricePoint: (point) => void;
      addOrder: (order: Order) => void;
    }
    ```
14. `src/stores/themeStore.ts`: `theme: 'light'|'dark'`, toggle, persist to localStorage

### FAZE 4: UI Primitives (`src/components/ui/`)
15. **Button.tsx** — variant: `primary|secondary|outline|ghost|danger`, size: `sm|md|lg`, loading state
16. **Input.tsx** — label, error message, icon prefix/suffix, disabled
17. **Card.tsx** — padding, hover effects, glassmorphism seçeneği
18. **Modal.tsx** — overlay, animation (framer-motion), close button
19. **Badge.tsx** — variant: `success|danger|warning|info|neutral`
20. **Spinner.tsx** — SVG animasyonlu loading indicator
21. **Table.tsx** — responsive, striped rows, hover
22. **Skeleton.tsx** — loading placeholder
23. **ThemeToggle.tsx** — Sun/Moon icon toggle, themeStore kullan
24. **Toast** — sonner `<Toaster />` Root layout'a ekle

### FAZE 5: Layout Components
25. **Navbar.tsx** (Public) — Logo, nav links (Home, Features), Login/Register butonları, mobile hamburger
26. **Sidebar.tsx** (Dashboard) — Logo, nav items (Dashboard, Trade, Wallet, Settings), user avatar, logout, collapsible (mobile)
27. **Topbar.tsx** (Dashboard) — Arama (dekoratif), bildirim ikonu, kullanıcı dropdown, theme toggle
28. **Footer.tsx** — Logo, linkler, sosyal medya, copyright
29. **MobileMenu.tsx** — Sheet/drawer tarzı mobil menü

### FAZE 6: Sayfalar

#### 30. Landing Page (`src/app/page.tsx`)
- `<Navbar />` + `<HeroSection />` + `<FeatureCard />`x4 + `<StatsSection />` + `<CTASection />` + `<Footer />`
- HeroSection: Büyük başlık, alt başlık, 2 CTA butonu (Kayıt Ol, Demo), sağda dekoratif gradient blob veya 3D grafik görseli
- FeatureCards: ikon + başlık + açıklama (4 adet grid)
- StatsSection: Animasyonlu sayaçlar (Kullanıcı, İşlem Hacmi, Uptime, vb.)
- **Responsive:** Mobilde tek kolon, desktop'ta grid

#### 31. Auth Layout (`src/app/(auth)/layout.tsx`)
- Desktop: sol %50 form, sağ %50 dekoratif gradient arka plan
- Mobil: sadece form alanı, tam genişlik

#### 32. Login Page (`src/app/(auth)/login/page.tsx`)
- `<LoginForm />`: react-hook-form + zod validasyon
- Email + Password input, "Giriş Yap" butonu
- "Hesabın yok mu? Kayıt Ol" linki
- Hata mesajı gösterimi (toast + inline)
- Submit → `authApi.login()` → `authStore.login()` → redirect `/dashboard`

#### 33. Register Page (`src/app/(auth)/register/page.tsx`)
- `<RegisterForm />`: Ad, Soyad, Email, Password (min 8 char)
- Validasyon: zod schema
- Submit → `authApi.register()` → `authStore.login()` → redirect `/dashboard`

#### 34. Dashboard Layout (`src/app/(dashboard)/layout.tsx`)
- `<Sidebar />` (sol 240px, collapsible) + `<Topbar />` (üst 64px) + `<main>` content alanı
- Auth guard: `useAuth` hook → token yoksa `/login` redirect
- Mobilde sidebar gizli, hamburger ile açılır

#### 35. Dashboard Page (`src/app/(dashboard)/dashboard/page.tsx`)
- **PortfolioCard:** Toplam USD bakiye + BTC karşılığı USD değeri (btc_balance × livePrice)
- **PriceCard:** BTCUSDT canlı fiyat (WebSocket), yeşil/kırmızı değişim animasyonu
- **MiniChart:** Son 50 fiyat noktası ile sparkline grafik (lightweight-charts)
- **QuickTradeWidget:** Basit BUY/SELL butonları + miktar input
- **RecentOrdersTable:** Son 5 emir (status badge'leri ile)
- Mount'ta: `walletStore.fetchWallet()` + WebSocket bağlantısı

#### 36. Trade Page (`src/app/(dashboard)/trade/page.tsx`)
- **Layout:** Desktop → sol %70 chart + sağ %30 order form. Mobil → üst chart, alt form (stacked)
- **PriceChart:** TradingView `lightweight-charts` createChart + addCandlestickSeries veya addLineSeries. WebSocket'ten gelen verilerle real-time güncelleme
- **TickerBar:** Üstte yatay bar — sembol, son fiyat, 24s değişim
- **OrderForm:**
  - BUY/SELL toggle (yeşil/kırmızı)
  - MARKET/LIMIT tab seçimi
  - Quantity input (number)
  - Price input (sadece LIMIT'te aktif)
  - Toplam hesap: quantity × price
  - "Emri Gönder" butonu → `ordersApi.placeOrder()` → toast success/error
  - Validasyon: side, type, quantity zorunlu; LIMIT'te price > 0 zorunlu
- **OrderHistory:** Tablo — ID, Side, Type, Quantity, Price, Status, Date

#### 37. Wallet Page (`src/app/(dashboard)/wallet/page.tsx`)
- **BalanceCards:** 2 kart — USD bakiye + BTC bakiye (ikon + rakam + küçük label)
- **DepositForm:** Modal veya inline — amount input + "Yatır" butonu → `walletApi.deposit(amount)`
- **WithdrawForm:** Modal veya inline — amount input + "Çek" butonu → `walletApi.withdraw(amount)`
  - Hata: "insufficient_balance" → toast uyarı
- **TransactionHistory:** Tablo — Tarih, Tür (DEPOSIT/WITHDRAWAL badge), Tutar, Durum

#### 38. Settings Page (`src/app/(dashboard)/settings/page.tsx`)
- Profil kartı: isim, email (read-only, authStore'dan)
- Tema toggle: Light/Dark switch
- Basit ve sade layout

#### 39. 404 Page (`src/app/not-found.tsx`)
- Büyük "404" tipografi
- "Sayfa bulunamadı" mesajı
- "Ana Sayfaya Dön" butonu

### FAZE 7: WebSocket Entegrasyonu
40. `src/lib/ws/binance.ts`:
```typescript
class BinanceWS {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;

  connect(onPrice: (price: string) => void) {
    this.ws = new WebSocket('wss://stream.binance.com:443/ws/btcusdt@trade');
    this.ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      onPrice(data.p); // price field
    };
    this.ws.onclose = () => this.reconnect(onPrice);
  }

  private reconnect(onPrice: (price: string) => void) {
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
    setTimeout(() => { this.reconnectAttempts++; this.connect(onPrice); }, delay);
  }

  disconnect() { this.ws?.close(); }
}
```

41. `src/hooks/useLivePrice.ts`:
```typescript
export function useLivePrice() {
  const { setLivePrice, addPricePoint } = useTradeStore();
  useEffect(() => {
    const ws = new BinanceWS();
    ws.connect((price) => {
      setLivePrice(price);
      addPricePoint({ time: Date.now() / 1000, value: parseFloat(price) });
    });
    return () => ws.disconnect();
  }, []);
}
```

### FAZE 8: Middleware & Route Protection
42. `src/middleware.ts`:
```typescript
import { NextRequest, NextResponse } from 'next/server';
const protectedRoutes = ['/dashboard', '/trade', '/wallet', '/settings'];
export function middleware(request: NextRequest) {
  const token = request.cookies.get('token')?.value;
  if (protectedRoutes.some(r => request.nextUrl.pathname.startsWith(r)) && !token) {
    return NextResponse.redirect(new URL('/login', request.url));
  }
}
```

### FAZE 9: Polish & Animations
43. Framer-motion page transitions (fade-in)
44. Skeleton loading states tüm veri-bağımlı bileşenlerde
45. Hover micro-animations (kartlar, butonlar)
46. Fiyat değişim animasyonu (yeşil flash up, kırmızı flash down)
47. Responsive son kontrol: 320px, 768px, 1280px, 1536px

---

## 7. Kritik Kurallar

1. **Tüm parasal değerler string olarak taşınır** — backend DECIMAL(18,8) kullanıyor, float kullanma
2. **JWT token her korumalı istekte gönderilmeli** — Axios interceptor ile otomatik
3. **401 hatası global handle** — interceptor'da logout + redirect
4. **WebSocket auto-reconnect zorunlu** — exponential backoff
5. **Dark mode tüm bileşenlerde desteklenmeli** — `dark:` prefix kullan
6. **Responsive tüm sayfalarda zorunlu** — mobile-first yaklaşım
7. **Form validasyon client-side yapılmalı** — zod schema + react-hook-form
8. **Loading states her API çağrısında gösterilmeli** — Skeleton veya Spinner
9. **Error states kullanıcıya gösterilmeli** — sonner toast
10. **Semantic HTML** — h1 her sayfada tek, aria-labels, proper heading hierarchy

---

## 8. Dosya Oluşturma Checklist

- [ ] `tailwind.config.ts` — extend colors, fonts, dark mode
- [ ] `globals.css` — CSS variables, base styles
- [ ] `src/types/domain.ts`
- [ ] `src/lib/utils/cn.ts`
- [ ] `src/lib/constants.ts`
- [ ] `src/lib/api/client.ts`
- [ ] `src/lib/api/auth.ts`
- [ ] `src/lib/api/wallet.ts`
- [ ] `src/lib/api/orders.ts`
- [ ] `src/lib/ws/binance.ts`
- [ ] `src/stores/authStore.ts`
- [ ] `src/stores/walletStore.ts`
- [ ] `src/stores/tradeStore.ts`
- [ ] `src/stores/themeStore.ts`
- [ ] `src/hooks/useAuth.ts`
- [ ] `src/hooks/useWallet.ts`
- [ ] `src/hooks/useLivePrice.ts`
- [ ] `src/hooks/useOrders.ts`
- [ ] `src/hooks/useTheme.ts`
- [ ] `src/components/ui/Button.tsx`
- [ ] `src/components/ui/Input.tsx`
- [ ] `src/components/ui/Card.tsx`
- [ ] `src/components/ui/Modal.tsx`
- [ ] `src/components/ui/Badge.tsx`
- [ ] `src/components/ui/Spinner.tsx`
- [ ] `src/components/ui/Table.tsx`
- [ ] `src/components/ui/Skeleton.tsx`
- [ ] `src/components/ui/ThemeToggle.tsx`
- [ ] `src/components/layout/Navbar.tsx`
- [ ] `src/components/layout/Sidebar.tsx`
- [ ] `src/components/layout/Topbar.tsx`
- [ ] `src/components/layout/Footer.tsx`
- [ ] `src/components/layout/MobileMenu.tsx`
- [ ] `src/components/landing/HeroSection.tsx`
- [ ] `src/components/landing/FeatureCard.tsx`
- [ ] `src/components/landing/StatsSection.tsx`
- [ ] `src/components/landing/CTASection.tsx`
- [ ] `src/components/auth/LoginForm.tsx`
- [ ] `src/components/auth/RegisterForm.tsx`
- [ ] `src/components/dashboard/PortfolioCard.tsx`
- [ ] `src/components/dashboard/PriceCard.tsx`
- [ ] `src/components/dashboard/QuickTradeWidget.tsx`
- [ ] `src/components/dashboard/RecentOrdersTable.tsx`
- [ ] `src/components/dashboard/MiniChart.tsx`
- [ ] `src/components/trade/PriceChart.tsx`
- [ ] `src/components/trade/OrderForm.tsx`
- [ ] `src/components/trade/TickerBar.tsx`
- [ ] `src/components/trade/OrderHistory.tsx`
- [ ] `src/components/wallet/BalanceCards.tsx`
- [ ] `src/components/wallet/DepositForm.tsx`
- [ ] `src/components/wallet/WithdrawForm.tsx`
- [ ] `src/components/wallet/TransactionHistory.tsx`
- [ ] `src/app/layout.tsx`
- [ ] `src/app/page.tsx`
- [ ] `src/app/not-found.tsx`
- [ ] `src/app/(auth)/layout.tsx`
- [ ] `src/app/(auth)/login/page.tsx`
- [ ] `src/app/(auth)/register/page.tsx`
- [ ] `src/app/(dashboard)/layout.tsx`
- [ ] `src/app/(dashboard)/dashboard/page.tsx`
- [ ] `src/app/(dashboard)/trade/page.tsx`
- [ ] `src/app/(dashboard)/wallet/page.tsx`
- [ ] `src/app/(dashboard)/settings/page.tsx`
- [ ] `src/middleware.ts`
