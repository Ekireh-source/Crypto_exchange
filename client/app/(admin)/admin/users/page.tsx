'use client';

import { useEffect, useState } from 'react';
import { adminService, AdminUser } from '@/feature/admin/admin.service';
import { useAppSelector } from '@/store/hooks';
import { toast } from 'sonner';

export default function AdminUsersPage() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(true);
  const { user: currentUser } = useAppSelector((state) => state.auth);

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      const data = await adminService.getUsers();
      setUsers(data);
    } catch (error) {
      console.error('Failed to fetch users:', error);
      toast.error('Failed to load users');
    } finally {
      setLoading(false);
    }
  };

  const handleRoleChange = async (userId: string, newRole: string) => {
    try {
      await adminService.updateUserRole(userId, newRole);
      toast.success('User role updated');
      fetchUsers();
    } catch (error) {
      console.error('Failed to update role:', error);
      toast.error('Failed to update role');
    }
  };

  if (loading) {
    return <div className="text-gray-400">Loading users...</div>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Manage Users</h1>
      
      <div className="bg-[#13151a] border border-[#22252e] rounded-2xl overflow-hidden shadow-xl">
        <div className="overflow-x-auto">
          <table className="w-full text-left text-sm">
            <thead className="bg-[#1a1d24] text-gray-400">
              <tr>
                <th className="px-6 py-4 font-medium">Email</th>
                <th className="px-6 py-4 font-medium">Role</th>
                <th className="px-6 py-4 font-medium">KYC Status</th>
                <th className="px-6 py-4 font-medium">Joined</th>
                <th className="px-6 py-4 font-medium text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#22252e]">
              {users.map((u) => (
                <tr key={u.id} className="hover:bg-[#1a1d24]/50 transition-colors">
                  <td className="px-6 py-4 text-white">
                    <div>{u.email}</div>
                    <div className="text-xs text-gray-500">{u.id}</div>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${
                      u.role === 'superadmin' ? 'bg-purple-500/10 text-purple-400' :
                      u.role === 'admin' ? 'bg-blue-500/10 text-blue-400' :
                      'bg-gray-500/10 text-gray-400'
                    }`}>
                      {u.role || 'user'}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${
                      u.kyc_status === 'verified' ? 'bg-green-500/10 text-green-400' :
                      'bg-yellow-500/10 text-yellow-400'
                    }`}>
                      {u.kyc_status}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-gray-400">
                    {new Date(u.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4 text-right">
                    {currentUser?.role === 'superadmin' && currentUser.id !== u.id && (
                      <select
                        className="bg-[#0a0b0d] border border-[#22252e] rounded-lg px-2 py-1 text-xs text-white focus:outline-none focus:border-blue-500"
                        value={u.role || 'user'}
                        onChange={(e) => handleRoleChange(u.id, e.target.value)}
                      >
                        <option value="user">User</option>
                        <option value="admin">Admin</option>
                        <option value="superadmin">Superadmin</option>
                      </select>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
