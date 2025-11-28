/**
 * Data item status enum
 */
export enum DataItemStatus {
  ACTIVE = 0,
  ARCHIVED = 1,
  DELETED = 2,
}

/**
 * Data item
 */
export interface DataItem {
  id: string;
  owner: string;
  data: string;
  metadata: Record<string, any>;
  hash: string;
  size: number;
  status: DataItemStatus;
  encrypted: boolean;
  createdAt: Date;
  updatedAt: Date;
  accessList?: string[];
  version: number;
}

/**
 * Register data parameters
 */
export interface RegisterDataParams {
  owner: string;
  data: string;
  metadata?: Record<string, any>;
  encrypted?: boolean;
  accessList?: string[];
}

/**
 * Update data parameters
 */
export interface UpdateDataParams {
  id: string;
  owner: string;
  data?: string;
  metadata?: Record<string, any>;
  accessList?: string[];
}

/**
 * Delete data parameters
 */
export interface DeleteDataParams {
  id: string;
  owner: string;
  permanent?: boolean;
}

/**
 * Data query filters
 */
export interface DataQueryFilters {
  owner?: string;
  status?: DataItemStatus;
  encrypted?: boolean;
  fromDate?: Date;
  toDate?: Date;
  tags?: string[];
  limit?: number;
  offset?: number;
}

/**
 * Data registry parameters
 */
export interface DataRegistryParams {
  maxDataSize: number;
  storageFee: string;
  updateFee: string;
  deleteFee: string;
  encryptionRequired: boolean;
  versioning: boolean;
  maxVersions: number;
}

/**
 * Data access grant
 */
export interface DataAccessGrant {
  dataId: string;
  grantee: string;
  grantor: string;
  permissions: string[];
  expiresAt?: Date;
  createdAt: Date;
}

/**
 * Data statistics
 */
export interface DataStats {
  totalItems: number;
  activeItems: number;
  archivedItems: number;
  totalSize: number;
  ownerCount: number;
  avgItemSize: number;
}
