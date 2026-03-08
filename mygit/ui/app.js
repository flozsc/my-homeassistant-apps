const API_BASE = '/api/v1';

class Router {
  constructor() {
    this.routes = {};
    window.addEventListener('hashchange', () => this.handleRoute());
  }

  add(path, handler) {
    this.routes[path] = handler;
  }

  navigate(path) {
    window.location.hash = path;
  }

  handleRoute() {
    const hash = window.location.hash.slice(1) || '/';
    const path = hash.split('?')[0];
    const query = this.parseQuery(hash);

    const handler = this.routes[path] || this.routes['*'];
    if (handler) {
      handler(query);
    }
  }

  parseQuery(hash) {
    const query = {};
    const qs = hash.split('?')[1];
    if (qs) {
      qs.split('&').forEach(param => {
        const [k, v] = param.split('=');
        query[k] = decodeURIComponent(v || '');
      });
    }
    return query;
  }

  getParam(name) {
    return this.handleRouteQuery ? this.handleRouteQuery[name] : null;
  }
}

const router = new Router();
let currentUser = null;

function el(id) {
  return document.getElementById(id);
}

function html(strings, ...values) {
  return strings.reduce((result, str, i) => {
    return result + str + (values[i] !== undefined ? values[i] : '');
  }, '');
}

async function api(endpoint, options = {}) {
  const token = localStorage.getItem('token');
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}${endpoint}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    localStorage.removeItem('token');
    currentUser = null;
    updateAuthUI();
    throw new Error('Unauthorized');
  }

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || 'Request failed');
  }

  return response.json();
}

function updateAuthUI() {
  const userInfo = el('user-info');
  const logoutBtn = el('logout-btn');
  const newRepoLink = el('nav-new-repo');

  if (currentUser) {
    userInfo.textContent = currentUser;
    logoutBtn.style.display = 'inline-flex';
    newRepoLink.style.display = 'inline';
  } else {
    userInfo.textContent = '';
    logoutBtn.style.display = 'none';
    newRepoLink.style.display = 'none';
  }
}

function showLogin() {
  el('content').innerHTML = html`
    <div class="login-container">
      <div class="modal">
        <h2 class="modal-title">Login to mygit</h2>
        <form id="login-form">
          <div class="form-group">
            <label class="form-label">Username</label>
            <input type="text" name="username" required>
          </div>
          <div class="form-group">
            <label class="form-label">Password</label>
            <input type="password" name="password" required>
          </div>
          <div id="login-error" class="form-error" style="display:none;"></div>
          <button type="submit" class="btn btn-primary btn-full">Login</button>
        </form>
      </div>
    </div>
  `;

  el('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const form = e.target;
    const username = form.username.value;
    const password = form.password.value;

    try {
      const authHeader = btoa(`${username}:${password}`);
      const response = await fetch(`${API_BASE}/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Basic ${authHeader}`,
        },
        body: JSON.stringify({ username, password }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Login failed');
      }

      const data = await response.json();
      localStorage.setItem('token', data.token);
      currentUser = username;
      updateAuthUI();
      router.navigate('/');
    } catch (err) {
      const errorEl = el('login-error');
      errorEl.textContent = err.message;
      errorEl.style.display = 'block';
    }
  });
}

async function renderRepoList() {
  el('content').innerHTML = '<div class="loading">Loading repositories...</div>';

  try {
    const repos = await api('/repos');

    if (repos.length === 0) {
      el('content').innerHTML = html`
        <div class="empty-state">
          <h2>No repositories yet</h2>
          <p>Create your first repository to get started.</p>
          <a href="#/new" class="btn btn-primary" style="margin-top:16px;">New Repository</a>
        </div>
      `;
      return;
    }

    el('content').innerHTML = html`
      <h1 class="page-title">Repositories</h1>
      <ul class="repo-list">
        ${repos.map(repo => html`
          <li class="repo-item">
            <div class="repo-item-header">
              <div>
                <div class="repo-name">
                  <a href="#/repo/${repo.name}">${repo.name}</a>
                </div>
                ${repo.description ? html`<div class="repo-description">${repo.description}</div>` : ''}
              </div>
              ${currentUser ? html`
                <div class="repo-actions">
                  <button class="btn btn-small btn-danger" onclick="deleteRepo('${repo.name}')">Delete</button>
                </div>
              ` : ''}
            </div>
            <div class="repo-meta">
              Last commit: ${repo.last_commit || 'none'} · Size: ${repo.size || '0 B'}
            </div>
          </li>
        `).join('')}
      </ul>
    `;
  } catch (err) {
    if (err.message === 'Unauthorized') {
      showLogin();
    } else {
      el('content').innerHTML = html`<div class="error">${err.message}</div>`;
    }
  }
}

async function renderNewRepo() {
  if (!currentUser) {
    router.navigate('/');
    return;
  }

  el('content').innerHTML = html`
    <h1 class="page-title">New Repository</h1>
    <form id="new-repo-form">
      <div class="form-group">
        <label class="form-label">Repository Name</label>
        <input type="text" name="name" pattern="^[a-zA-Z0-9_-]+$" required placeholder="my-repo">
        <small style="color:var(--text-secondary)">Only letters, numbers, underscores, and hyphens</small>
      </div>
      <div class="form-group">
        <label class="form-label">Description (optional)</label>
        <input type="text" name="description" placeholder="A short description">
      </div>
      <div id="new-repo-error" class="form-error" style="display:none;"></div>
      <button type="submit" class="btn btn-primary">Create Repository</button>
    </form>
  `;

  el('new-repo-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const form = e.target;
    const name = form.name.value.trim();
    const description = form.description.value.trim();

    try {
      await api('/repos', {
        method: 'POST',
        body: JSON.stringify({ name, description }),
      });
      router.navigate('/');
    } catch (err) {
      const errorEl = el('new-repo-error');
      errorEl.textContent = err.message;
      errorEl.style.display = 'block';
    }
  });
}

async function deleteRepo(name) {
  if (!confirm(`Delete repository "${name}"? This cannot be undone.`)) {
    return;
  }

  try {
    await api(`/repos/${name}`, { method: 'DELETE' });
    renderRepoList();
  } catch (err) {
    alert(err.message);
  }
}

async function renderRepo(params) {
  const name = params.repo;
  el('content').innerHTML = '<div class="loading">Loading repository...</div>';

  try {
    const [repo, branches, contributors] = await Promise.all([
      api(`/repos/${name}`),
      api(`/repos/${name}/branches`),
      api(`/repos/${name}/contributors`),
    ]);

    const commits = await api(`/repos/${name}/commits?limit=5`);

    el('content').innerHTML = html`
      <h1 class="page-title">${repo.name}</h1>

      <div class="tabs">
        <button class="tab active" data-tab="code">Code</button>
        <button class="tab" data-tab="commits">Commits</button>
        <button class="tab" data-tab="branches">Branches</button>
        <button class="tab" data-tab="contributors">Contributors</button>
      </div>

      <div id="tab-code" class="tab-content active">
        ${renderFileTree(repo.name, repo.default_branch)}
      </div>

      <div id="tab-commits" class="tab-content">
        ${renderCommitList(commits, repo.name)}
      </div>

      <div id="tab-branches" class="tab-content">
        ${renderBranchList(branches, repo.default_branch)}
      </div>

      <div id="tab-contributors" class="tab-content">
        ${renderContributorList(contributors)}
      </div>
    `;

    setupTabs(name);
  } catch (err) {
    if (err.message === 'Unauthorized') {
      showLogin();
    } else {
      el('content').innerHTML = html`<div class="error">${err.message}</div>`;
    }
  }
}

function renderFileTree(repoName, branch) {
  return html`
    <div id="file-browser">
      <div class="loading">Loading files...</div>
    </div>
  `;
}

function renderCommitList(commits, repoName) {
  if (!commits || commits.length === 0) {
    return '<p class="empty-state">No commits yet</p>';
  }

  return html`
    <ul class="commit-list">
      ${commits.map(commit => html`
        <li class="commit-item">
          <div class="commit-info">
            <div class="commit-message">
              <a href="#/repo/${repoName}/commit/${commit.sha}">${commit.message}</a>
            </div>
            <div class="commit-meta">
              <span class="commit-sha">${commit.sha.substring(0, 7)}</span>
              · ${commit.author} · ${formatDate(commit.date)}
            </div>
          </div>
        </li>
      `).join('')}
    </ul>
  `;
}

function renderBranchList(branches, currentBranch) {
  if (!branches || branches.length === 0) {
    return '<p class="empty-state">No branches</p>';
  }

  return html`
    <div class="branch-list">
      ${branches.map(branch => html`
        <span class="branch-badge ${branch === currentBranch ? 'current' : ''}">
          ${branch === currentBranch ? '✓ ' : ''}${branch}
        </span>
      `).join('')}
    </div>
  `;
}

function renderContributorList(contributors) {
  if (!contributors || contributors.length === 0) {
    return '<p class="empty-state">No contributors</p>';
  }

  return html`
    <ul class="repo-list">
      ${contributors.map(c => html`
        <li class="repo-item">
          <div class="repo-name">${c.name || c.email}</div>
          <div class="repo-meta">${c.commits} commits</div>
        </li>
      `).join('')}
    </ul>
  `;
}

function setupTabs(repoName) {
  const tabs = document.querySelectorAll('.tab');
  tabs.forEach(tab => {
    tab.addEventListener('click', () => {
      tabs.forEach(t => t.classList.remove('active'));
      tab.classList.add('active');

      document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
      const tabId = tab.dataset.tab;
      el(`tab-${tabId}`).classList.add('active');

      if (tabId === 'code') {
        loadFileTree(repoName, 'main');
      }
    });
  });
}

async function loadFileTree(repoName, branch, path = '') {
  const fileBrowser = el('file-browser');
  if (!fileBrowser) return;

  const endpoint = path
    ? `/repos/${repoName}/tree/${branch}/${path}`
    : `/repos/${repoName}/tree/${branch}`;

  fileBrowser.innerHTML = '<div class="loading">Loading files...</div>';

  try {
    const files = await api(endpoint);
    renderFiles(repoName, branch, path, files);
  } catch (err) {
    fileBrowser.innerHTML = html`<div class="error">${err.message}</div>`;
  }
}

function renderFiles(repoName, branch, path, files) {
  const fileBrowser = el('file-browser');
  if (!fileBrowser) return;

  const parts = path ? path.split('/') : [];
  const breadcrumbs = html`
    <div class="breadcrumb">
      <a href="#/repo/${repoName}">${repoName}</a>
      ${parts.map((part, i) => {
        const prefix = parts.slice(0, i + 1).join('/');
        return html`
          <span class="breadcrumb-sep">/</span>
          <a href="#/repo/${repoName}/tree/${branch}/${prefix}">${part}</a>
        `;
      })}
    </div>
  `;

  if (path) {
    const parentPath = parts.slice(0, -1).join('/');
    const parentBreadcrumb = html`
      <div class="breadcrumb">
        <a href="#/repo/${repoName}">${repoName}</a>
        ${parts.slice(0, -1).map((part, i) => {
          const prefix = parts.slice(0, i + 1).join('/');
          return html`
            <span class="breadcrumb-sep">/</span>
            <a href="#/repo/${repoName}/tree/${branch}/${prefix}">${part}</a>
          `;
        })}
        <span class="breadcrumb-sep">/</span>
        <span class="breadcrumb-current">${parts[parts.length - 1]}</span>
      </div>
    `;
  }

  fileBrowser.innerHTML = html`
    ${path ? breadcrumbs : ''}
    <ul class="file-list">
      ${path ? html`
        <li class="file-item">
          <span class="file-icon">📁</span>
          <span class="file-name">
            <a href="#/repo/${repoName}/tree/${branch}/${parentPath}">..</a>
          </span>
        </li>
      ` : ''}
      ${files.map(file => html`
        <li class="file-item">
          <span class="file-icon">${file.type === 'tree' ? '📁' : '📄'}</span>
          <span class="file-name">
            ${file.type === 'tree'
              ? html`<a href="#/repo/${repoName}/tree/${branch}/${path ? path + '/' + file.name : file.name}">${file.name}</a>`
              : html`<a href="#/repo/${repoName}/blob/${path ? path + '/' + file.name : file.name}">${file.name}</a>`
            }
          </span>
          <span class="file-size">${file.size || ''}</span>
        </li>
      `).join('')}
    </ul>
  `;
}

function formatDate(dateStr) {
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

async function checkAuth() {
  try {
    const user = await api('/auth/me');
    currentUser = user.username;
    updateAuthUI();
  } catch (err) {
    currentUser = null;
    updateAuthUI();
  }
}

function initRouter() {
  router.add('/', renderRepoList);
  router.add('/new', renderNewRepo);
  router.add('/repo/:repo', (params) => renderRepo({ repo: params.repo }));
  router.add('/settings', renderSettings);
}

document.addEventListener('DOMContentLoaded', async () => {
  initRouter();

  el('logout-btn').addEventListener('click', () => {
    localStorage.removeItem('token');
    currentUser = null;
    updateAuthUI();
    router.navigate('/');
  });

  await checkAuth();
  router.handleRoute();
});

function renderSettings() {
  el('content').innerHTML = html`
    <div class="settings-page">
      <h1 class="page-title">Settings</h1>

      <div class="settings-section">
        <h3>API Keys</h3>
        <p>Generate API keys for programmatic access to your repositories.</p>

        <div id="api-keys-section">
          <!-- API keys will be rendered here -->
        </div>

        <form id="create-api-key-form">
          <div class="form-group">
            <label for="api-key-scopes">Scopes</label>
            <select id="api-key-scopes" name="scopes" multiple>
              <option value="read">Read</option>
              <option value="write">Write</option>
              <option value="admin">Admin</option>
            </select>
          </div>
          <button type="submit" class="btn btn-primary">Generate API Key</button>
        </form>
      </div>

      <div class="settings-section" id="user-management-section" style="display:none;">
        <h3>User Management</h3>
        <table class="table" id="users-table">
          <thead>
            <tr>
              <th>Username</th>
              <th>Scopes</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody id="users-tbody">
            <!-- Filled by JS -->
          </tbody>
        </table>
        <form id="create-user-form">
          <div class="form-group">
            <label for="new-username">Username</label>
            <input type="text" id="new-username" name="username" required>
          </div>
          <div class="form-group">
            <label for="new-password">Password</label>
            <input type="password" id="new-password" name="password" required>
          </div>
          <div class="form-group">
            <label for="new-scopes">Scopes</label>
            <select id="new-scopes" name="scopes" multiple>
              <option value="read">Read</option>
              <option value="write">Write</option>
              <option value="admin">Admin</option>
            </select>
          </div>
          <button type="submit" class="btn btn-primary">Create User</button>
        </form>
      </div>

      <div class="settings-section" id="ssh-key-section">
        <h3>SSH Keys</h3>
        <p>Manage SSH keys for Git operations over SSH.</p>

        <form id="add-ssh-key-form">
          <div class="form-group">
            <label for="ssh-key">Public Key</label>
            <textarea id="ssh-key" name="ssh_key" rows="5" required></textarea>
          </div>
          <button type="submit" class="btn btn-primary">Add SSH Key</button>
        </form>

        <table class="table" id="ssh-keys-table" style="display:none;">
          <thead>
            <tr>
              <th>Fingerprint</th>
              <th>Comment</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody id="ssh-keys-tbody">
            <!-- Filled by JS -->
          </tbody>
        </table>
        <div class="empty-state" id="no-ssh-keys-msg">No SSH keys yet.</div>
      </div>
    </div>
  `;

  initSettings();
}

function initSettings() {
  // Load current user info to determine admin status
  api('/auth/me')
    .then(user => {
      // Show user management section only for admin
      if (user && user.username) {
        // Attempt to load users list; if forbidden, we are not admin
        api('/users')
          .then(data => {
            document.getElementById('user-management-section').style.display = 'block';
            renderUserList(data.users);
          })
          .catch(() => {
            // Not admin or error; hide section
            document.getElementById('user-management-section').style.display = 'none';
          });
      }
    })
    .catch(() => {
      // Not authenticated; hide admin sections
      document.getElementById('user-management-section').style.display = 'none';
    });

  // Load SSH keys for current user
  api('/ssh-keys')
    .then(keys => {
      const table = document.getElementById('ssh-keys-table');
      const tbody = el('ssh-keys-tbody');
      const noMsg = el('no-ssh-keys-msg');
      if (keys.length > 0) {
        table.style.display = 'table';
        noMsg.style.display = 'none';
        tbody.innerHTML = keys.map(k => `
          <tr>
            <td>${k.Fingerprint}</td>
            <td>${k.Comment}</td>
            <td><button class="button small" data-fp="${k.Fingerprint}">Remove</button></td>
          </tr>`).join('');
        // Attach delete handlers
        tbody.querySelectorAll('button').forEach(btn => {
          btn.addEventListener('click', () => {
            const fp = btn.getAttribute('data-fp');
            api(`/ssh-keys/${fp}`, { method: 'DELETE' })
              .then(() => initSettings())
              .catch(err => alert(err.message));
          });
        });
      } else {
        table.style.display = 'none';
        noMsg.style.display = 'block';
      }
    })
    .catch(() => {
      // Not auth or error
    });

  // Bind create user form
  const createUserForm = el('create-user-form');
  if (createUserForm) {
    createUserForm.addEventListener('submit', e => {
      e.preventDefault();
      const username = el('new-username').value.trim();
      const password = el('new-password').value.trim();
      const scopes = Array.from(el('new-scopes').selectedOptions).map(o => o.value);
      api('/users', { method: 'POST', body: JSON.stringify({ username, password, scopes }) })
        .then(() => initSettings())
        .catch(err => alert(err.message));
    });
  }

  // Bind add SSH key form
  const addKeyForm = el('add-ssh-key-form');
  if (addKeyForm) {
    addKeyForm.addEventListener('submit', e => {
      e.preventDefault();
      const key = el('ssh-key').value.trim();
      api('/ssh-keys', { method: 'POST', body: JSON.stringify({ key }) })
        .then(() => initSettings())
        .catch(err => alert(err.message));
    });
  }

  // Bind create API key form
  const createApiKeyForm = el('create-api-key-form');
  if (createApiKeyForm) {
    createApiKeyForm.addEventListener('submit', e => {
      e.preventDefault();
      const scopes = Array.from(el('api-key-scopes').selectedOptions).map(o => o.value);
      // For now, just show a message - API key generation would need a backend endpoint
      alert('API key generation is not yet implemented');
    });
  }
}

function renderUserList(users) {
  const tbody = el('users-tbody');
  tbody.innerHTML = users.map(u => `
    <tr>
      <td>${u.Username}</td>
      <td>${u.Scopes.join(', ')}</td>
      <td><button class="button small" data-username="${u.Username}">Delete</button></td>
    </tr>`).join('');
  tbody.querySelectorAll('button').forEach(btn => {
    btn.addEventListener('click', () => {
      const uname = btn.getAttribute('data-username');
      if (confirm('Delete user ' + uname + '?')) {
        api(`/users/${uname}`, { method: 'DELETE' })
          .then(() => initSettings())
          .catch(err => alert(err.message));
      }
    });
  });
}

