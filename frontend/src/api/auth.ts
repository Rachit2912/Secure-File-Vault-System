import { postReq, getReq } from "./index";

// types for authentication request payloads :
export type SignupPayload = {
  username: string;
  email?: string;
  password: string;
};
export type LoginPayload = { username: string; password: string };

// user signup :
export async function apiSignup(p: SignupPayload) {
  return postReq("/api/signup", p);
}

// user login :
export async function apiLogin(p: LoginPayload) {
  return postReq("/api/login", p);
}

// user logout (clears cookies) :
export async function apiLogout() {
  return postReq("/api/logout", {});
}

// refresh current user session & authenticating it :
export async function apiMe() {
  return getReq("/api/me");
}
