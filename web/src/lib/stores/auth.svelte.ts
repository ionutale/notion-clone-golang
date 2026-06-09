import { api } from '$lib/api';
import type { User } from '$lib/types';

class AuthStore {
  user = $state<User | null>(null);
  accessToken = $state<string | null>(null);
  loading = $state(true);

  async login(email: string, password: string) {
    const res: any = await api.request('POST', '/auth/login', { email, password });
    this.user = res.user;
    this.accessToken = res.access_token;
  }

  async signup(email: string, password: string, name: string) {
    const res: any = await api.request('POST', '/auth/signup', { email, password, name });
    this.user = res.user;
    this.accessToken = res.access_token;
  }

  async logout() {
    await api.request('POST', '/auth/logout');
    this.user = null;
    this.accessToken = null;
  }

  async refresh() {
    try {
      const res: any = await api.requestInner('POST', '/auth/refresh');
      this.accessToken = res.access_token;
      this.user = res.user;
    } catch {
      this.user = null;
      this.accessToken = null;
    }
  }

  async check() {
    this.loading = true;
    await this.refresh();
    this.loading = false;
  }
}

export const authStore = new AuthStore();
