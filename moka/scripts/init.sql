DROP DATABASE IF EXISTS `moka_dev`;

CREATE DATABASE IF NOT EXISTS `moka_dev`;

-- ==================== 用户与组织架构 ====================
CREATE TABLE `users` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `username` VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
  `password` VARCHAR(255) NOT NULL COMMENT '密码(bcrypt)',
  `real_name` VARCHAR(100) COMMENT '真实姓名',
  `email` VARCHAR(100),
  `phone` VARCHAR(20),
  `dept_id` BIGINT DEFAULT 0 COMMENT '部门ID（无外键）',
  `status` TINYINT DEFAULT 1 COMMENT '状态: 1启用 0禁用',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX `idx_dept_id` (`dept_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

CREATE TABLE `departments` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `name` VARCHAR(100) NOT NULL COMMENT '部门名称',
  `parent_id` BIGINT DEFAULT 0 COMMENT '父部门ID（无外键）',
  `level` INT DEFAULT 1 COMMENT '层级',
  `path` VARCHAR(500) COMMENT '路径: 1/2/3/',
  `leader_id` BIGINT COMMENT '负责人ID（无外键）',
  `sort` INT DEFAULT 0,
  `status` TINYINT DEFAULT 1,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX `idx_parent_id` (`parent_id`),
  INDEX `idx_path` (`path`(255)),
  INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='部门表';

-- ==================== RBAC 核心表 ====================
CREATE TABLE `roles` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `code` VARCHAR(50) NOT NULL UNIQUE COMMENT '角色编码',
  `name` VARCHAR(100) NOT NULL COMMENT '角色名称',
  `parent_id` BIGINT DEFAULT 0 COMMENT '父角色ID（无外键）',
  `level` INT DEFAULT 1 COMMENT '角色层级',
  `data_scope` VARCHAR(20) DEFAULT 'self' COMMENT '数据范围',
  `remark` VARCHAR(500),
  `sort` INT DEFAULT 0,
  `status` TINYINT DEFAULT 1,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX `idx_parent_id` (`parent_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

CREATE TABLE `user_roles` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `user_id` BIGINT NOT NULL COMMENT '用户ID（无外键）',
  `role_id` BIGINT NOT NULL COMMENT '角色ID（无外键）',
  `expire_at` DATETIME DEFAULT NULL COMMENT '过期时间',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_role_id` (`role_id`),
  INDEX `idx_expire_at` (`expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

CREATE TABLE `permissions` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `type` VARCHAR(20) NOT NULL COMMENT '权限类型: api/menu/button',
  `code` VARCHAR(100) NOT NULL COMMENT '权限编码',
  `name` VARCHAR(100) NOT NULL COMMENT '权限名称',
  `resource` VARCHAR(255) NOT NULL COMMENT '资源标识（精确路径）',
  `action` VARCHAR(50) NOT NULL COMMENT '操作: GET/POST/PUT/DELETE/view/click',
  `parent_id` BIGINT DEFAULT 0 COMMENT '父级ID（无外键）',
  `path` VARCHAR(500) COMMENT '路径（菜单）',
  `icon` VARCHAR(100) COMMENT '图标',
  `component` VARCHAR(255) COMMENT '组件路径',
  `sort` INT DEFAULT 0,
  `status` TINYINT DEFAULT 1,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_resource_action_type` (`resource`, `action`, `type`),
  INDEX `idx_type` (`type`),
  INDEX `idx_parent_id` (`parent_id`),
  INDEX `idx_status` (`status`),
  INDEX `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限资源表';

CREATE TABLE `role_permissions` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `role_id` BIGINT NOT NULL COMMENT '角色ID（无外键）',
  `permission_id` BIGINT NOT NULL COMMENT '权限ID（无外键）',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_role_permission` (`role_id`, `permission_id`),
  INDEX `idx_role_id` (`role_id`),
  INDEX `idx_permission_id` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';

CREATE TABLE `role_data_scopes` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `role_id` BIGINT NOT NULL COMMENT '角色ID（无外键）',
  `dept_id` BIGINT NOT NULL COMMENT '部门ID（无外键）',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY `uk_role_dept` (`role_id`, `dept_id`),
  INDEX `idx_role_id` (`role_id`),
  INDEX `idx_dept_id` (`dept_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色数据权限表';

-- 审计日志表
CREATE TABLE `permission_logs` (
  `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
  `operator_id` BIGINT NOT NULL COMMENT '操作人ID',
  `operator_name` VARCHAR(100) COMMENT '操作人姓名',
  `operation` VARCHAR(50) NOT NULL COMMENT '操作类型',
  `target_type` VARCHAR(20) NOT NULL COMMENT '目标类型',
  `target_id` BIGINT NOT NULL COMMENT '目标ID',
  `resource_type` VARCHAR(20) COMMENT '资源类型',
  `resource_id` BIGINT COMMENT '资源ID',
  `detail` TEXT COMMENT '详细信息(JSON)',
  `ip` VARCHAR(50),
  `user_agent` VARCHAR(500),
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  INDEX `idx_operator_id` (`operator_id`),
  INDEX `idx_target` (`target_type`, `target_id`),
  INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限操作日志';