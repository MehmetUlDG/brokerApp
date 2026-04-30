# TradeOff - Kripto Para Borsası Frontend

TradeOff, modern, güvenilir ve yüksek performanslı bir kripto para borsası arayüzüdür. Hem yeni başlayanlar hem de profesyonel trader'lar için tasarlanmış olup, Binance WebSocket altyapısı sayesinde gerçek zamanlı piyasa verileri ve kesintisiz işlem imkanı sunar.

![TradeOff Demo](./public/demo.webp)

## 🚀 Özellikler

- **Modern ve Profesyonel Arayüz:** Kullanıcı dostu, "Light" ve "Dark" tema destekli, tamamen duyarlı (responsive) tasarım.
- **Gerçek Zamanlı Piyasa Verileri:** Binance WebSocket entegrasyonu ile BTC/USDT (ve yakında daha fazlası) için milisaniyelik gecikmesiz canlı fiyat güncellemeleri.
- **Gelişmiş Grafikler:** `lightweight-charts` ile profesyonel ve etkileşimli fiyat trend grafikleri.
- **Korumalı Rotalar (Protected Routes):** Kullanıcı oturumu gerektiren sayfalara Next.js middleware JWT kontrolleriyle güvenli erişim.
- **Cüzdan Yönetimi:** Bakiye görüntüleme, işlem geçmişi ve Stripe entegrasyonuna sahip altyapı ile para yatırma / çekme simülasyonları.
- **Gelişmiş Form Doğrulama:** `react-hook-form` ve `zod` kullanılarak hata toleranslı ve hızlı form yönetimi.
- **Global State:** Zustand ile performanslı ve hafif global state (Theme, Auth, Wallet, Trade) yönetimi.

## 🛠 Kullanılan Teknolojiler

- **Core:** [Next.js 14+](https://nextjs.org/) & [React 18](https://react.dev/)
- **Dil:** TypeScript
- **Stil & Tasarım Sistemi:** Tailwind CSS v4, CSS Variables, Lucide React (İkonlar)
- **State Yönetimi:** [Zustand](https://github.com/pmndrs/zustand)
- **Veri Çekme & API:** Axios
- **Form & Doğrulama:** React Hook Form & Zod
- **Grafikler:** [Lightweight Charts](https://tradingview.github.io/lightweight-charts/) (TradingView)
- **Tarih Formatlama:** date-fns
- **Ödeme & Cüzdan:** Stripe (React Stripe.js)

## 📁 Proje Yapısı

```text
src/
├── app/               # Next.js App Router yapısı (Sayfalar, Layoutlar)
│   ├── (auth)/        # Korumalı olmayan giriş/kayıt sayfaları
│   └── (dashboard)/   # Korumalı kontrol paneli, cüzdan ve alım/satım sayfaları
├── components/        # Yeniden kullanılabilir React bileşenleri
│   ├── auth/          # Kimlik doğrulama formları
│   ├── dashboard/     # Ana panel widgetları ve grafikleri
│   ├── landing/       # Ana sayfa bileşenleri
│   ├── layout/        # Navbar, Sidebar, Footer vb.
│   ├── trade/         # Emir defteri, fiyat grafiği, alım-satım formu
│   ├── ui/            # Button, Input, Card vb. temel UI elementleri
│   └── wallet/        # Cüzdan bakiyesi ve işlem formları
├── hooks/             # Özel React Hook'ları (useAuth, useLivePrice, useWallet vb.)
├── lib/               # Utility fonksiyonlar, sabitler ve API servisleri
├── stores/            # Zustand global state dosyaları
└── types/             # TypeScript domain interface ve type tanımlamaları
```

## ⚙️ Kurulum ve Çalıştırma

### Gereksinimler
- Node.js 18.x veya üzeri
- npm (veya yarn/pnpm)

### Kurulum Adımları

1. Repoyu bilgisayarınıza klonlayın ve dizine gidin:
   ```bash
   cd brokerApp/frontend
   ```

2. Bağımlılıkları yükleyin:
   ```bash
   npm install
   ```

3. Çevresel değişkenleri (Environment Variables) ayarlayın. Ana dizinde bir `.env.local` dosyası oluşturun ve aşağıdaki değişkenleri tanımlayın:
   ```env
   NEXT_PUBLIC_API_URL=http://localhost:8080   # Backend API adresiniz
   NEXT_PUBLIC_WS_URL=wss://stream.binance.com:443/ws/btcusdt@trade
   NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_... # Stripe test key
   ```

4. Geliştirme sunucusunu başlatın:
   ```bash
   npm run dev
   ```

5. Uygulamayı incelemek için tarayıcınızda [http://localhost:3000](http://localhost:3000) adresine gidin.

## 📦 Build ve Production Ortamı

Uygulamayı üretim (production) ortamı için derlemek ve çalıştırmak için:

```bash
npm run build
npm run start
```

---

*Geliştirme süreci boyunca backend (mikroservis mimarisi) ile tam uyumlu çalışacak şekilde REST API kontratlarına göre tasarlanmış ve optimize edilmiştir.*
