-- 03-action-lifecycle: Supply-chain demo data
-- Two tables that mirror a typical ERP setup.
-- eval_purchase_orders.supplier_status is denormalised from suppliers
-- so the action condition can filter directly on the PO object type.

DROP TABLE IF EXISTS eval_purchase_orders;
DROP TABLE IF EXISTS eval_suppliers;

CREATE TABLE eval_suppliers (
  id          VARCHAR(20)  NOT NULL PRIMARY KEY,
  name        VARCHAR(100) NOT NULL,
  status      VARCHAR(20)  NOT NULL COMMENT 'active | at_risk | suspended',
  since       DATE         NOT NULL,
  score       INT          NOT NULL COMMENT '0-100 quality score'
) DEFAULT CHARSET=utf8mb4 COMMENT='供应商主数据';

CREATE TABLE eval_purchase_orders (
  id              VARCHAR(20)  NOT NULL PRIMARY KEY,
  supplier_id     VARCHAR(20)  NOT NULL,
  supplier_status VARCHAR(20)  NOT NULL COMMENT 'denormalised from eval_suppliers.status',
  material        VARCHAR(100) NOT NULL,
  quantity        INT          NOT NULL,
  due_date        DATE         NOT NULL,
  status          VARCHAR(20)  NOT NULL COMMENT 'pending | confirmed | overdue'
) DEFAULT CHARSET=utf8mb4 COMMENT='采购单';

INSERT INTO eval_suppliers VALUES
  ('SUP-001', '华东电子有限公司', 'active',   '2024-01-15', 95),
  ('SUP-002', '北方机械制造厂',   'at_risk',  '2024-03-01', 62),
  ('SUP-003', '深圳精密零件',     'active',   '2023-11-20', 88),
  ('SUP-004', '西南原材料供应',   'at_risk',  '2024-02-28', 55),
  ('SUP-005', '东莞组装工厂',     'active',   '2024-01-05', 91);

INSERT INTO eval_purchase_orders VALUES
  ('PO-2024-001','SUP-001','active',  '电容器 100μF',    5000,'2024-04-15','confirmed'),
  ('PO-2024-002','SUP-002','at_risk', '钢材型材 Q235',   2000,'2024-04-10','pending'),
  ('PO-2024-003','SUP-002','at_risk', '螺栓 M8×40',    10000,'2024-04-20','pending'),
  ('PO-2024-004','SUP-003','active',  '铝合金外壳',       300,'2024-04-18','confirmed'),
  ('PO-2024-005','SUP-004','at_risk', '橡胶密封圈',      8000,'2024-04-12','pending'),
  ('PO-2024-006','SUP-004','at_risk', '聚氨酯泡沫',      1500,'2024-04-25','pending'),
  ('PO-2024-007','SUP-005','active',  'PCB 主板',         200,'2024-04-22','confirmed'),
  ('PO-2024-008','SUP-001','active',  '电阻 10kΩ',      20000,'2024-04-30','pending'),
  ('PO-2024-009','SUP-003','active',  '散热片',            500,'2024-04-16','confirmed'),
  ('PO-2024-010','SUP-005','active',  '显示屏组件',        150,'2024-04-28','pending');
