module.exports = {
  apps: [{
    name: 'cli-proxy-api-plus',
    script: './CLIProxyAPIPlus',
    cwd: '/Users/caolin/Desktop/projects/CLIProxyAPIPlus',
    instances: 1,
    autorestart: true,
    watch: false,
    max_memory_restart: '500M',
    env: {
      TZ: 'Asia/Shanghai'
    },
    error_file: './logs/pm2-error.log',
    out_file: './logs/pm2-out.log',
    log_date_format: 'YYYY-MM-DD HH:mm:ss Z',
    merge_logs: true,
    min_uptime: '10s',
    max_restarts: 10,
    restart_delay: 4000
  }]
};
