export function Footer() {
  return (
    <footer className="border-t border-[var(--border)] bg-[var(--surface)] py-8 mt-auto">
      <div className="container mx-auto px-4 text-center text-sm text-[var(--text-muted)]">
        &copy; {new Date().getFullYear()} TradeOff. Tüm hakları saklıdır.
      </div>
    </footer>
  );
}
