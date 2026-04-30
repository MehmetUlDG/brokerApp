'use client';

import { Card } from '@/components/ui/Card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { useTradeStore } from '@/stores/tradeStore';
import { format } from 'date-fns';

export function OrderHistory() {
  const orders = useTradeStore((state) => state.orders);

  return (
    <Card className="overflow-hidden">
      <div className="border-b border-[var(--border)] p-4">
        <h3 className="font-bold text-[var(--text-primary)]">Emir Geçmişi</h3>
      </div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Zaman</TableHead>
            <TableHead>Sembol</TableHead>
            <TableHead>Yön</TableHead>
            <TableHead>Tür</TableHead>
            <TableHead>Fiyat</TableHead>
            <TableHead>Miktar</TableHead>
            <TableHead>Durum</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {orders.length === 0 ? (
            <TableRow>
              <TableCell colSpan={7} className="text-center text-[var(--text-muted)] py-12">
                Emir geçmişiniz bulunmuyor.
              </TableCell>
            </TableRow>
          ) : (
            orders.map((order) => (
              <TableRow key={order.id}>
                <TableCell className="text-[var(--text-secondary)]">
                  {format(new Date(order.created_at), 'dd MMM yyyy HH:mm:ss')}
                </TableCell>
                <TableCell className="font-medium text-[var(--text-primary)]">{order.symbol}</TableCell>
                <TableCell>
                  <span className={order.side === 'BUY' ? 'text-[var(--success)] font-medium' : 'text-[var(--danger)] font-medium'}>
                    {order.side === 'BUY' ? 'Alış' : 'Satış'}
                  </span>
                </TableCell>
                <TableCell className="text-[var(--text-secondary)]">{order.type}</TableCell>
                <TableCell className="font-mono text-[var(--text-primary)]">${parseFloat(order.price).toLocaleString()}</TableCell>
                <TableCell className="font-mono text-[var(--text-primary)]">{order.quantity}</TableCell>
                <TableCell>
                  <Badge
                    variant={
                      order.status === 'COMPLETED' ? 'success' :
                      order.status === 'PENDING' ? 'warning' :
                      order.status === 'FAILED' ? 'danger' : 'neutral'
                    }
                  >
                    {order.status}
                  </Badge>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </Card>
  );
}
