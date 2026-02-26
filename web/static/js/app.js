/*
 * Copyright (C) 2025 Russell Li (xiaoxinmm)
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

// 违纪管理系统 - 前端核心

const App = {
  user: null,

  init() {
    // cookie 是 HttpOnly 的，JS 不需要读取
    // 浏览器会自动带上 cookie
  },

  async api(url, options = {}) {
    const headers = { ...options.headers };

    if (options.json) {
      headers['Content-Type'] = 'application/json';
      options.body = JSON.stringify(options.json);
      delete options.json;
    }

    const res = await fetch(url, {
      ...options,
      headers,
      credentials: 'same-origin'
    });

    if (res.status === 401) {
      window.location.href = '/login';
      return null;
    }

    return res;
  },

  async apiJSON(url, options = {}) {
    const res = await this.api(url, options);
    if (!res) return null;
    return res.json();
  },

  toast(message, type) {
    type = type || 'success';
    var container = document.querySelector('.toast-box');
    if (!container) {
      container = document.createElement('div');
      container.className = 'toast-box';
      document.body.appendChild(container);
    }

    var el = document.createElement('div');
    el.className = 'toast ' + type;

    var prefix = '';
    if (type === 'error') prefix = '[错误] ';
    else if (type === 'warning') prefix = '[提示] ';

    el.textContent = prefix + message;
    container.appendChild(el);

    setTimeout(function () {
      el.style.opacity = '0';
      setTimeout(function () { el.remove(); }, 300);
    }, 3000);
  },

  escapeHtml(str) {
    if (!str) return '';
    var d = document.createElement('div');
    d.textContent = str;
    return d.innerHTML;
  },

  formatDate(s) {
    var d = new Date(s);
    var y = d.getFullYear();
    var m = ('0' + (d.getMonth() + 1)).slice(-2);
    var day = ('0' + d.getDate()).slice(-2);
    return y + '-' + m + '-' + day;
  },

  formatDateTime(s) {
    var d = new Date(s);
    var y = d.getFullYear();
    var m = ('0' + (d.getMonth() + 1)).slice(-2);
    var day = ('0' + d.getDate()).slice(-2);
    var h = ('0' + d.getHours()).slice(-2);
    var min = ('0' + d.getMinutes()).slice(-2);
    return y + '-' + m + '-' + day + ' ' + h + ':' + min;
  },

  showModal(id) {
    var el = document.getElementById(id);
    if (el) el.classList.add('active');
  },

  hideModal(id) {
    var el = document.getElementById(id);
    if (el) el.classList.remove('active');
  },

  async logout() {
    await this.api('/api/logout', { method: 'POST' });
    window.location.href = '/';
  },

  async checkAuth() {
    var data = await this.apiJSON('/api/me');
    if (data && data.user) {
      this.user = data.user;
      return data.user;
    }
    // apiJSON 在 401 时已经跳转了
    return null;
  }
};

App.init();
