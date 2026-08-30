/**
 * Hand-written domain types that mirror the backend DTOs. Kept alongside the
 * generated `schema.ts` because ProTable columns / form fields read from these
 * shapes directly (the generated types are path-keyed and less ergonomic).
 */

export type Captcha = {
  captchaId: string;
  captchaBase64: string;
};

export type LoginRequest = {
  username: string;
  password: string;
  captchaId: string;
  captchaCode: string;
};

export type LoginResponse = {
  accessToken: string;
  refreshToken?: string;
  tokenType: string;
  expiresIn: number;
};

export type CurrentUser = {
  userId: string;
  username: string;
  nickname: string;
  avatar: string;
  gender: number;
  mobile: string;
  email: string;
  deptName: string;
  createTime: string;
  canSwitchTenant: boolean;
  roles: string[];
  perms: string[];
};

export type UserProfileUpdate = Partial<
  Pick<CurrentUser, 'nickname' | 'avatar' | 'mobile' | 'email' | 'gender'>
>;

/** 当前用户自助修改密码（PUT /users/password） */
export type ChangePasswordRequest = {
  oldPassword: string;
  newPassword: string;
};

export type UserOption = {
  value: string;
  label: string;
};

export type User = {
  id: string;
  username: string;
  nickname: string;
  gender: number;
  deptId: string;
  deptName: string;
  avatar: string;
  mobile: string;
  status: number;
  email: string;
  createTime: string;
  updateTime: string;
  roleIds: string[];
};

export type UserForm = {
  id?: string;
  username: string;
  nickname: string;
  password?: string;
  mobile?: string;
  gender?: number;
  avatar?: string;
  email?: string;
  status?: number;
  deptId?: string;
  roleIds?: string[];
};

export type UserQuery = {
  username?: string;
  nickname?: string;
  status?: number;
  deptId?: string;
  pageNum?: number;
  pageSize?: number;
};

export type Role = {
  id: string;
  name: string;
  code: string;
  sort: number;
  status: number;
  dataScope?: number;
  createTime: string;
  updateTime: string;
};

export type RoleForm = {
  id?: string;
  name: string;
  code: string;
  sort?: number;
  status?: number;
  dataScope?: number;
};

export type RoleOption = { value: string; label: string };

export type MenuTypeCode = 'M' | 'C' | 'B';

export type MenuNode = {
  id: string;
  parentId: string;
  name: string;
  type: MenuTypeCode;
  routeName?: string;
  routePath?: string;
  component?: string;
  perm?: string;
  alwaysShow?: number;
  keepAlive?: number;
  visible: number;
  sort: number;
  icon?: string;
  redirect?: string;
  children?: MenuNode[];
};

export type MenuForm = {
  id?: string;
  parentId: string;
  name: string;
  type: MenuTypeCode;
  routeName?: string;
  routePath?: string;
  component?: string;
  perm?: string;
  alwaysShow?: number;
  keepAlive?: number;
  visible?: number;
  sort?: number;
  icon?: string;
  redirect?: string;
};

export type MenuOption = {
  value: string;
  label: string;
  children?: MenuOption[];
};

export type Dept = {
  id: string;
  name: string;
  code: string;
  parentId: string;
  sort: number;
  status: number;
  createTime: string;
  updateTime: string;
  children?: Dept[];
};

export type DeptForm = {
  id?: string;
  name: string;
  code: string;
  parentId: string;
  sort?: number;
  status?: number;
};

export type DeptOption = {
  value: string;
  label: string;
  children?: DeptOption[];
};

export type Dict = {
  id: string;
  dictCode: string;
  name: string;
  status: number;
  remark: string;
  createTime: string;
};

export type DictForm = {
  id?: string;
  dictCode: string;
  name: string;
  status?: number;
  remark?: string;
};

export type DictItem = {
  id: string;
  dictCode: string;
  label: string;
  value: string;
  tagType: string;
  sort: number;
  status: number;
  remark: string;
  createTime: string;
};

export type DictItemOption = {
  label: string;
  value: string;
  tagType?: string;
};

export type DictItemForm = {
  id?: string;
  dictCode?: string;
  label: string;
  value: string;
  tagType?: string;
  sort?: number;
  status?: number;
  remark?: string;
};

export type FileUploadResult = {
  name: string;
  url: string;
};

export type Config = {
  id: string;
  configName: string;
  configKey: string;
  configValue: string;
  remark: string;
  createTime: string;
};

export type ConfigForm = {
  id?: string;
  configName: string;
  configKey: string;
  configValue: string;
  remark?: string;
};

export type LogRow = {
  id: string;
  module: string;
  requestMethod: string;
  content: string;
  requestUri: string;
  ip: string;
  province: string;
  city: string;
  executionTime: number;
  browser: string;
  browserVersion: string;
  os: string;
  createBy: string;
  createTime: string;
};

export type Notice = {
  id: string;
  title: string;
  type: number;
  level: string;
  targetType: number;
  publishStatus: number;
  publishTime: string | null;
  publisherName: string;
  createTime: string;
};

export type NoticeForm = {
  id?: string;
  title: string;
  content?: string;
  type: number;
  level: string;
  targetType: number;
  targetUserIds?: string[];
};

export type NoticeDetail = {
  id: string;
  title: string;
  content: string;
  type: number;
  level: string;
  publisherId: string;
  publisherName: string;
  publishTime: string | null;
};

export type TaskStatus =
  | 'queued'
  | 'processing'
  | 'completed'
  | 'error'
  | 'canceled'
  | 'suspending';

export type TaskStats = {
  busyWorkers: number;
  successTasks: number;
  failureTasks: number;
  submittedTasks: number;
  suspendingTasks: number;
  queuedCount: number;
  processingCount: number;
  completedCount: number;
  errorCount: number;
  canceledCount: number;
};

export type TaskType = { label: string; value: string };

export type TaskRow = {
  id: number;
  type: string;
  status: TaskStatus;
  correlationId: string;
  ownerId: string;
  retryCount: number;
  executedDuration: number;
  error: string;
  errorHistory: string;
  resumeTime: number;
  createdAt: string;
  updatedAt: string;
};

export type DownloadStatus =
  | 'downloading'
  | 'seeding'
  | 'completed'
  | 'error'
  | 'queued'
  | 'canceled';

export type DownloadStats = {
  downloadingCount: number;
  seedingCount: number;
  completedCount: number;
  errorCount: number;
  totalCount: number;
};

export type DownloaderInfo = { label: string; value: string };

export type DownloadRow = {
  id: number;
  taskId: string;
  hash: string;
  name: string;
  url: string;
  downloader: string;
  status: DownloadStatus;
  total: number;
  downloaded: number;
  downloadSpeed: number;
  uploaded: number;
  uploadSpeed: number;
  savePath: string;
  errorMessage: string;
  progress: number;
  createdAt: string;
  updatedAt: string;
};

export type DownloadFile = {
  index: number;
  name: string;
  size: number;
  progress: number;
  selected: boolean;
};

export type DownloadDetail = DownloadRow & { files: DownloadFile[] };

export type CreateDownloadRequest = {
  url: string;
  downloader?: string;
  options?: Record<string, unknown>;
};

export type UserNotice = {
  id: string;
  noticeId: string;
  title: string;
  type: number;
  level: string;
  publishTime: string | null;
  isRead: number;
};

export type RouteItem = {
  path: string;
  name?: string;
  component?: string;
  redirect?: string;
  icon?: string;
  hideInMenu?: boolean;
  keepAlive?: boolean;
  perms?: string[];
  children?: RouteItem[];
  type?: 'dir' | 'menu' | 'button';
};

// --- 多租户 / 会员 ---------------------------------------------------------

export type Tenant = {
  id: string;
  code: string;
  name: string;
  status: number;
  createTime?: string;
  updateTime?: string;
};

export type TenantForm = {
  id?: string;
  code: string;
  name: string;
  status?: number;
};

export type MemberRow = {
  id: string;
  tenantId: string;
  username: string;
  nickname: string;
  email: string;
  mobile: string;
  status: number;
  lastLoginTime?: string;
  createTime?: string;
};
