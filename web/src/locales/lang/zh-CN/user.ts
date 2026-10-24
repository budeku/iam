export default {
  field: {
    id: 'ID',
    username: '用户名',
    password: '密码',
    uuid: 'UUID',
    email: '邮箱',
    status: '状态',
    createdAt: '创建时间',
    updatedAt: '更新时间',
    LastLoginAt: '最近登录时间',
  },
  status: {
    active: '启用',
    disabled: '禁用',
  },
  searchPlaceholder: '搜索用户名/邮箱',
  deleteConfirm: '确定要删除用户 {name} 吗？此操作不可撤销',
  form: {
    createTitle: '新建用户',
    editTitle: '编辑用户',
    usernameRequired: '请输入用户名',
    emailRequired: '请输入邮箱',
    emailInvalid: '邮箱格式不正确',
    passwordRequired: '请输入密码',
    passwordEditHint: '留空表示不修改密码',
    superAdminTip: '超级管理员账号不可修改或操作',
  },
}
