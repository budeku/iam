export default {
  field: {
    kid: 'Kid',
    algorithm: '算法',
    status: '状态',
    activatedAt: '启用时间',
    retireAt: '停用时间',
    createdAt: '创建时间',
    appId: '应用',
    appName: '应用',
    publicKey: '公钥',
    privateKey: '私钥',
  },
  form: {
    createTitle: '新建密钥对',
    detailTitle: '密钥对详情',
    algorithmRequired: '请选择算法',
    statusRequired: '请选择状态',
  },
  downgrade: '降级',
  downgradeConfirm: '降级后该密钥将不再用于新签名，仅用于验证旧令牌，确定继续吗？',
  retire: '停用',
  retireConfirm: '停用后该密钥将完全失效，确定继续吗？',
}
