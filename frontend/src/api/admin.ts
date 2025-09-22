import { getReq, postReq } from "./index";
import type { FilterValues } from "../components/filters/Filters";
import type { FileMeta } from "./files";

// fetch all files (admin only, supports filters) :
export async function listAdminFiles(filters?: Partial<FilterValues>): Promise<{
  files: FileMeta[];
  originalSize: number;
  dedupSize: number;
  saveSize: number;
}> {
  let url = "/api/adminFiles";

  // 1. if filters exist, convert to query params :
  if (filters) {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value) params.append(key, value.toString());
    });
    url += "?" + params.toString();
  }

  // 2. call backend and reporting errors + response:
  const res = await getReq(url);
  if (!res) throw new Error("Failed to fetch admin files");
  return res;
}

// promote an user to admin role :
export async function makeAdmin(username: string) {
  const res = await postReq("/api/makeAdmin", { username });
  if (!res) throw new Error("Failed to make user admin");
  return res;
}

// demote an admin back to normal user :
export async function makeUser(username: string) {
  const res = await postReq("/api/makeUser", { username });
  if (!res) throw new Error("Failed to make user normal user");
  return res;
}
