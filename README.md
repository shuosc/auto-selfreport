# Auto Self Report

## Usage

### Crontab
```bash
$ crontab -e

# Auto Self Report
30 6 * * * curl -L -o /tmp/report https://github.com/shuosc/auto-selfreport/releases/latest/download/auto-selfreport_linux_amd64 && chmod +x /tmp/report && /tmp/report -u your-username -p your-password -e your-email
```

### GitHub Action

After forking, add a file named `report.yml` to .github/workflows with following contents:

```yaml
name: Auto Self Report
on: 
  schedule:
    - cron: "30 22 * * *"
jobs:

  report:
    name: Self report
    runs-on: ubuntu-latest
    steps:

    - name: Get executable
      run: |
        curl -L -o auto-self-report https://github.com/shuosc/auto-selfreport/releases/latest/download/auto-selfreport_linux_amd64
        chmod +x auto-self-report
        
    - name: Run job
      run: ./auto-self-report -u ${{ secrets.SHU_USERNAME }} -p ${{ secrets.SHU_PASSWORD }} -e ${{ secrets.EMAIL }}

```

(The time of cron above is an UTC time)

Then, add secrets at your own repository's <a href="../../settings/secrets">Settings-Secrets</a>

- `SHU_USERNAME`: 一卡通账号
- `SHU_PASSWORD`: 一卡通密码
- `EMAIL`: 用于接收通知的邮箱

This job will run at 6:30 a.m. every day.