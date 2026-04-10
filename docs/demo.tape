# sonacli demo tape — regenerates docs/demo.gif
#
# Prerequisites: make, Go 1.26+, podman, vhs, jq, curl, a cached
# SonarScanner CLI (see scripts/seed-sample-project.sh for one).
#
# Regenerate:
#   1. make build
#   2. ./scripts/local-sonarqube.sh start
#   3. Create and analyze the demo project (once per fresh SonarQube):
#        curl -fsS -u admin:SonacliAdmin1@ -X POST \
#          "http://127.0.0.1:9000/api/projects/create" \
#          --data-urlencode "project=payments-api" \
#          --data-urlencode "name=Payments API"
#        SCANNER=tests/bin/sample-project-tools/sonar-scanner-*/bin/sonar-scanner
#        SCAN_TOKEN=$(curl -fsS -u admin:SonacliAdmin1@ -X POST \
#          "http://127.0.0.1:9000/api/user_tokens/generate" \
#          --data-urlencode "name=demo-scan-$(date +%s)" | jq -r .token)
#        ( cd docs/_demo-source && \
#          SONAR_HOST_URL=http://127.0.0.1:9000 SONAR_TOKEN=$SCAN_TOKEN \
#          "$PWD/../../$SCANNER" )
#   4. export TOKEN=$(curl -fsS -u admin:SonacliAdmin1@ -X POST \
#        "http://127.0.0.1:9000/api/user_tokens/generate" \
#        --data-urlencode "name=demo-$(date +%s)" | jq -r .token)
#   5. export ISSUE_KEY=$(curl -fsS -u admin:SonacliAdmin1@ \
#        "http://127.0.0.1:9000/api/issues/search?componentKeys=payments-api&impactSeverities=HIGH" \
#        | jq -r '.issues[0].key')
#   6. vhs docs/demo.tape

Output docs/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 760
Set Theme "Dracula"
Set TypingSpeed 30ms
Set PlaybackSpeed 1.0
Set Padding 25
Set WindowBar Colorful

Hide
Type 'export PATH="$PWD:$PATH" HOME=/tmp/sonacli-demo-home PS1="$ "'
Enter
Type 'mkdir -p $HOME && rm -rf $HOME/.sonacli && cd /tmp && clear'
Enter
Show

Sleep 600ms

# Scene 1 — Connect sonacli to a SonarQube instance.
Type "sonacli auth setup -s http://127.0.0.1:9000 -t $TOKEN"
Sleep 300ms
Enter
Sleep 1500ms

# Scene 2 — Browse projects as pretty JSON.
Type "sonacli project list --pretty"
Sleep 300ms
Enter
Sleep 1800ms

# Scene 3 — sonacli speaks JSON: compose with jq to reshape issues.
Type "sonacli issue list payments-api | jq '.issues[] | {rule, sev: .impacts[0].severity, file: .component, line, message}'"
Sleep 300ms
Enter
Sleep 2200ms

# Scene 4 — Filter by impact severity.
Type "sonacli issue list payments-api --severity HIGH | jq '.issues[] | {rule, file: .component, line, message}'"
Sleep 300ms
Enter
Sleep 2200ms

# Scene 5 — Paste a SonarQube web URL; sonacli extracts the issue key for you.
Type@6ms 'sonacli issue show "http://127.0.0.1:9000/project/issues?id=payments-api&issues=$ISSUE_KEY" | jq "{rule, file: .component, line, impact: .impacts[0], message}"'
Sleep 300ms
Enter
Sleep 2500ms

Sleep 500ms
