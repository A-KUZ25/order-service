package mysql

const QueryUnpaidOrdersBase = `
SELECT o.order_id
FROM tbl_order o
WHERE o.tenant_id = ?
  AND o.active = 1
  AND o.status_id = ?
`
