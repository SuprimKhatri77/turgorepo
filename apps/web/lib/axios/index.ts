import axios from "axios";
import type { AxiosInstance } from "axios";
import { attachResponseInterceptor } from "./response-interceptor";

const api: AxiosInstance = axios.create({
  baseURL: `${process.env.NEXT_PUBLIC_API_URL}/api/v1`,
  withCredentials: true,
});

attachResponseInterceptor(api);

export default api;
