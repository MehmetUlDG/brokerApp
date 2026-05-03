'use client';

import { Card } from '@/components/ui/Card';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/Table';
import { Badge } from '@/components/ui/Badge';
import { Skeleton } from '@/components/ui/Skeleton';
import { useOrders } from '@/hooks/useOrders';
import { format } from 'date-fns';

export function OrderHistory() {
  const { orders, ordersLoading } = useOrders();

  return (
    <Card className="overflow-hidden">
      <div className="border-b border-[var(--border)] p-4 flex items-center justify-between">
        <h3 className="font-bold text-[var(--text-primary)]">Emir Geçmişi</h3>
        {ordersLoading && (
          <span className="text-xs text-[var(--text-muted)] animate-pulse">Güncelleniyor...</span>
        )}
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
          {ordersLoading && orders.length === 0 ? (
            <>
              {[...Array(4)].map((_, i) => (
                <TableRow key={i}>
                  {[...Array(7)].map((_, j) => (
                    <TableCell key={j}><Skeleton className="h-4 w-full" /></TableCell>
                  ))}
                </TableRow>
              ))}
            </>
          ) : orders.length === 0 ? (
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
