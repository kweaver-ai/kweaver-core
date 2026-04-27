-- supply_chain sample data for KWeaver 01-db-to-qa example
-- Fictional smart-home company — all names are fabricated
--
-- The database must already exist. run.sh connects with: mysql ... "$DB_NAME" < seed.sql
-- (Do not put CREATE DATABASE / USE here — the MySQL user may only have schema rights on that DB.)

DROP TABLE IF EXISTS `erp_material_bom`;
CREATE TABLE `erp_material_bom` (
  `seq_no` varchar(255) DEFAULT NULL,
  `bom_version` varchar(255) DEFAULT NULL,
  `bom_material_code` varchar(255) DEFAULT NULL,
  `bom_level` varchar(255) DEFAULT NULL,
  `material_code` varchar(255) DEFAULT NULL,
  `material_name` varchar(255) DEFAULT NULL,
  `usage_numerator` varchar(255) DEFAULT NULL,
  `usage_denominator` varchar(255) DEFAULT NULL,
  `standard_usage` varchar(255) DEFAULT NULL,
  `variable_loss_rate` varchar(255) DEFAULT NULL,
  `demand_quantity` varchar(255) DEFAULT NULL,
  `alt_group_no` varchar(255) DEFAULT NULL,
  `alt_part` varchar(255) DEFAULT NULL,
  `alt_priority` varchar(255) DEFAULT NULL,
  `alt_method` varchar(255) DEFAULT NULL,
  `parent_material_code` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `erp_purchase_order`;
CREATE TABLE `erp_purchase_order` (
  `id` bigint NOT NULL COMMENT '采购订单主ID',
  `billno` varchar(64) DEFAULT NULL COMMENT '采购订单编号',
  `biztime` date DEFAULT NULL COMMENT '业务日期',
  `org_name` varchar(128) DEFAULT NULL COMMENT '采购组织名称',
  `billtype_name` varchar(64) DEFAULT NULL COMMENT '单据类型名称',
  `biztype_name` varchar(64) DEFAULT NULL COMMENT '业务类型名称',
  `supplier_number` varchar(32) DEFAULT NULL COMMENT '供应商编号',
  `supplier_name` varchar(128) DEFAULT NULL COMMENT '供应商名称',
  `operatorname` varchar(32) DEFAULT NULL COMMENT '操作人姓名',
  `huid_xycg_operatorname` varchar(32) DEFAULT NULL COMMENT '寻源采购员姓名',
  `createtime` datetime DEFAULT NULL COMMENT '订单创建时间',
  `auditdate` datetime DEFAULT NULL COMMENT '审核时间',
  `entry_id` bigint DEFAULT NULL COMMENT '订单明细行ID',
  `material_number` varchar(32) DEFAULT NULL COMMENT '物料编码',
  `material_name` varchar(128) DEFAULT NULL COMMENT '物料名称',
  `qty` decimal(12,2) DEFAULT NULL COMMENT '采购数量',
  `tax_price` decimal(10,2) DEFAULT NULL COMMENT '含税单价',
  `taxrate` decimal(5,2) DEFAULT NULL COMMENT '税率',
  `deliverdate` date DEFAULT NULL COMMENT '要求交货日期',
  `srcbillentryid` bigint DEFAULT NULL COMMENT '来源单据明细ID',
  `srcbillid` bigint DEFAULT NULL COMMENT '来源单据主ID',
  `srcbillnumber` varchar(32) DEFAULT NULL COMMENT '来源单据编号',
  `invqty` decimal(12,2) DEFAULT NULL COMMENT '入库数量',
  `returnqty` decimal(12,2) DEFAULT NULL COMMENT '退货数量',
  `actqty` decimal(12,2) DEFAULT NULL COMMENT '实际数量',
  `rowclosestatus_title` varchar(32) DEFAULT NULL COMMENT '行关闭状态名称',
  `rowterminatestatus_title` varchar(32) DEFAULT NULL COMMENT '行终止状态名称'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- Sample data: erp_material_bom (10 rows — one product BOM tree)
-- ---------------------------------------------------------------------------
INSERT INTO `erp_material_bom` VALUES
('1','2026-01-15','FP-900001','0','FP-900001','智能温控器总成',NULL,NULL,'1','0','1','0','','0','',''),
('2','2026-01-15','FP-900001','1','MD-200010','主控电路板_PCBA','1','1','1','0','1','0','','0','','FP-900001'),
('3','2026-01-15','FP-900001','1','MD-200020','温度传感器模组','2','1','2','0','2','0','','0','','FP-900001'),
('4','2026-01-15','FP-900001','1','MD-200030','LCD显示屏_3.5寸','1','1','1','0','1','0','','0','','FP-900001'),
('5','2026-01-15','FP-900001','2','CP-100001','贴片电阻_10kΩ','20','1','20','0','20','0','','0','','MD-200010'),
('6','2026-01-15','FP-900001','2','CP-100002','贴片电容_100nF','15','1','15','0','15','0','','0','','MD-200010'),
('7','2026-01-15','FP-900001','2','CP-100003','微控制器芯片_ARM','1','1','1','0','1','0','','0','','MD-200010'),
('8','2026-01-15','FP-900001','2','CP-100004','Wi-Fi通信模块','1','1','1','0','1','1','','0','替代','MD-200010'),
('9','2026-01-15','FP-900001','2','CP-100005','蓝牙通信模块','1','1','1','0','1','1','√','1','替代','MD-200010'),
('10','2026-01-15','FP-900001','2','CP-100006','NTC热敏电阻','2','1','2','0','2','0','','0','','MD-200020');

-- ---------------------------------------------------------------------------
-- Sample data: erp_purchase_order (10 rows)
-- ---------------------------------------------------------------------------
INSERT INTO `erp_purchase_order` VALUES
(1001,'PO-260301-0001','2026-03-01','星河智能科技有限公司','生产采购订单','物料类采购','S-0001','深圳芯联微电子有限公司','张伟','王芳','2026-03-01 09:30:00','2026-03-01 14:00:00',2001,'CP-100003','微控制器芯片_ARM',500.00,12.50,13.00,'2026-03-20',3001,4001,'PR-260228-001',0.00,0.00,0.00,NULL,NULL),
(1002,'PO-260301-0002','2026-03-01','星河智能科技有限公司','生产采购订单','物料类采购','S-0002','苏州晶元电子科技有限公司','张伟','王芳','2026-03-01 09:45:00','2026-03-01 14:10:00',2002,'CP-100001','贴片电阻_10kΩ',10000.00,0.03,13.00,'2026-03-15',3002,4001,'PR-260228-001',10000.00,0.00,10000.00,'已关闭',NULL),
(1003,'PO-260301-0003','2026-03-01','星河智能科技有限公司','生产采购订单','物料类采购','S-0002','苏州晶元电子科技有限公司','张伟','王芳','2026-03-01 09:50:00','2026-03-01 14:15:00',2003,'CP-100002','贴片电容_100nF',8000.00,0.05,13.00,'2026-03-15',3003,4001,'PR-260228-001',8000.00,0.00,8000.00,'已关闭',NULL),
(1004,'PO-260302-0001','2026-03-02','星河智能科技有限公司','生产采购订单','物料类采购','S-0003','杭州博远传感技术有限公司','李娜','王芳','2026-03-02 10:00:00','2026-03-02 15:30:00',2004,'MD-200020','温度传感器模组',1000.00,8.80,13.00,'2026-03-25',3004,4002,'PR-260301-002',200.00,0.00,200.00,NULL,NULL),
(1005,'PO-260302-0002','2026-03-02','星河智能科技有限公司','生产采购订单','物料类采购','S-0004','东莞光显电子有限公司','李娜','赵敏','2026-03-02 10:30:00',NULL,2005,'MD-200030','LCD显示屏_3.5寸',500.00,35.00,13.00,'2026-03-28',3005,4002,'PR-260301-002',0.00,0.00,0.00,NULL,NULL),
(1006,'PO-260303-0001','2026-03-03','星河智能科技有限公司','生产采购订单','物料类采购','S-0005','广州联讯无线科技有限公司','李娜','赵敏','2026-03-03 11:00:00','2026-03-03 16:00:00',2006,'CP-100004','Wi-Fi通信模块',500.00,15.00,13.00,'2026-03-30',3006,4003,'PR-260302-003',500.00,0.00,500.00,'已关闭',NULL),
(1007,'PO-260303-0002','2026-03-03','星河智能科技有限公司','研发采购订单','物料类采购','S-0006','无锡恒芯半导体有限公司','张伟','赵敏','2026-03-03 14:00:00',NULL,2007,'CP-100005','蓝牙通信模块',200.00,18.00,13.00,'2026-04-05',3007,4004,'PR-260303-004',0.00,0.00,0.00,NULL,NULL),
(1008,'PO-260305-0001','2026-03-05','星河智能科技有限公司','生产采购订单','物料类采购','S-0003','杭州博远传感技术有限公司','张伟','王芳','2026-03-05 09:00:00','2026-03-05 11:00:00',2008,'CP-100006','NTC热敏电阻',2000.00,1.20,13.00,'2026-03-20',3008,4005,'PR-260304-005',2000.00,0.00,2000.00,'已关闭',NULL),
(1009,'PO-260305-0002','2026-03-05','星河智能科技有限公司','生产采购订单','物料类采购','S-0001','深圳芯联微电子有限公司','李娜','王芳','2026-03-05 09:30:00','2026-03-06 10:00:00',2009,'CP-100003','微控制器芯片_ARM',300.00,12.30,13.00,'2026-03-25',3009,4005,'PR-260304-005',300.00,50.00,250.00,NULL,NULL),
(1010,'PO-260306-0001','2026-03-06','星河智能科技有限公司','生产采购订单','物料类采购','S-0004','东莞光显电子有限公司','李娜','赵敏','2026-03-06 08:30:00',NULL,2010,'MD-200030','LCD显示屏_3.5寸',800.00,34.50,13.00,'2026-04-01',3010,4006,'PR-260305-006',0.00,0.00,0.00,NULL,NULL);
