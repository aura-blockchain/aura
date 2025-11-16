# Incident Communication Plan

## Purpose

This document defines the communication strategy and procedures for incident response, ensuring timely, accurate, and coordinated information sharing with all stakeholders during security incidents.

## Communication Principles

1. **Transparency**: Be honest and clear about what happened
2. **Timeliness**: Communicate quickly, update frequently
3. **Accuracy**: Verify information before sharing
4. **Consistency**: Maintain unified messaging across channels
5. **Empathy**: Acknowledge user concerns and impact

## Stakeholder Categories

### Internal Stakeholders

#### Executive Leadership
- **Who**: CEO, CTO, CFO, Board of Directors
- **When**: Immediately for Critical/High incidents
- **How**: Phone call + Secure email + Emergency meeting
- **Information**: Full technical details, impact assessment, response plan

#### Technical Teams
- **Who**: Engineering, DevOps, Security teams
- **When**: Immediately upon incident detection
- **How**: Slack #incident-response, PagerDuty, Conference bridge
- **Information**: Technical details, action items, real-time updates

#### Operations Teams
- **Who**: Customer support, Community managers
- **When**: Within 30 minutes of incident
- **How**: Internal briefing document, Slack #ops-briefing
- **Information**: Prepared responses, FAQ, escalation procedures

### External Stakeholders

#### Users
- **Who**: All active users, token holders
- **When**: Within 1 hour for Critical, 4 hours for High
- **How**: Email, In-app notifications, Status page
- **Information**: Impact summary, recommended actions, next update time

#### Validators
- **Who**: All validator operators
- **When**: Immediately for any chain-level incident
- **How**: Validator private channel, Direct SMS
- **Information**: Technical details, coordination requirements, action items

#### Partners & Exchanges
- **Who**: Integration partners, exchanges listing AURA
- **When**: Within 2 hours of incident
- **How**: Email + API webhooks
- **Information**: Trading impact, deposit/withdrawal status, technical updates

#### Media & Public
- **Who**: Journalists, crypto media, general public
- **When**: After internal stakeholders informed
- **How**: Press release, Blog post, Twitter
- **Information**: Approved public statement, factual summary

## Communication Channels

### Primary Channels

#### 1. Status Page
- **URL**: https://status.aura-network.io
- **Purpose**: Real-time service status
- **Update Frequency**: Every 30 min (Critical), Every 2 hours (High)
- **Audience**: All users

```yaml
# Status page configuration
status_page:
  provider: "Statuspage.io"
  page_id: "aura-network"
  components:
    - name: "Blockchain RPC"
      status: operational|degraded|partial_outage|major_outage
    - name: "API Services"
      status: operational|degraded|partial_outage|major_outage
    - name: "Explorer"
      status: operational|degraded|partial_outage|major_outage
    - name: "Validators"
      status: operational|degraded|partial_outage|major_outage
```

#### 2. Email Notifications
- **Lists**: all-users@aura.io, validators@aura.io, partners@aura.io
- **Templates**: Pre-approved templates for each severity
- **Approval**: Security Lead (Critical), Ops Lead (Others)

```python
# Automated email notification
def send_incident_notification(incident):
    template = get_template(incident.severity)
    recipients = get_recipients(incident.affected_systems)

    email = {
        'subject': f'[{incident.severity.upper()}] Aura Network Incident',
        'to': recipients,
        'body': template.render(incident=incident),
        'reply_to': 'security@aura-network.io'
    }

    send_email(email)
    log_communication(incident.id, 'email', recipients)
```

#### 3. Social Media
- **Twitter**: @AuraNetwork
- **Discord**: #announcements
- **Telegram**: @AuraOfficial
- **Frequency**: Major updates only
- **Approval Required**: Communications Lead

```markdown
# Twitter Template - Critical Incident
🚨 Aura Network Alert

We have detected a [severity] incident affecting [systems].
Current status: [status]

✅ User funds: [safe/at risk/investigating]
✅ Chain status: [operational/paused/degraded]

Updates: https://status.aura-network.io
Incident ID: [INC-XXX]

#AuraNetwork
```

#### 4. In-App Notifications
```javascript
// In-app notification banner
{
  "type": "incident",
  "severity": "critical",
  "message": "We are currently investigating a security incident. Some services may be unavailable.",
  "cta": {
    "text": "Learn More",
    "url": "https://status.aura-network.io"
  },
  "dismissible": false
}
```

### Secondary Channels

#### 5. Blog Posts
- **URL**: https://blog.aura-network.io
- **Purpose**: Detailed technical explanations
- **Timing**: After incident resolution
- **Approval**: Executive team

#### 6. Validator Private Channel
- **Platform**: Encrypted Telegram group
- **Members**: Verified validator operators only
- **Purpose**: Coordination, technical updates
- **Response Time**: < 15 minutes

#### 7. Partner Portal
- **URL**: https://partners.aura-network.io
- **Purpose**: Real-time API status, integration updates
- **Authentication**: Partner API keys

## Communication Templates

### Critical Incident - Initial Notification

```
Subject: [CRITICAL] Aura Network Security Incident

Dear Aura Community,

We are currently responding to a critical security incident affecting the Aura Network.

CURRENT STATUS:
- Incident detected at: [TIME] UTC
- Affected systems: [SYSTEMS]
- Chain status: [OPERATIONAL/PAUSED]
- User funds: [SAFE/INVESTIGATING]

IMMEDIATE ACTIONS TAKEN:
- [Action 1]
- [Action 2]
- [Action 3]

WHAT THIS MEANS FOR YOU:
[Impact description - be specific about what users can/cannot do]

WHAT WE'RE DOING:
Our security team is actively investigating and working on a resolution. We have:
- Identified the root cause
- Implemented containment measures
- Coordinated with validators
- [Other actions]

NEXT STEPS:
- Next update: [TIME] UTC (in X hours)
- Status page: https://status.aura-network.io
- Incident ID: [INC-XXX]

We understand this is concerning and appreciate your patience. User security is our top priority.

If you have urgent questions, contact: security@aura-network.io

Aura Security Team
[TIMESTAMP]

---
Track this incident: https://status.aura-network.io/incidents/[INC-XXX]
```

### High Incident - Update Notification

```
Subject: [UPDATE] Aura Network Incident [INC-XXX]

Status Update - [TIMESTAMP]

CURRENT STATUS: [INVESTIGATING/CONTAINED/RESOLVING/MONITORING]

PROGRESS SINCE LAST UPDATE:
- [Update 1]
- [Update 2]
- [Update 3]

IMPACT UPDATE:
- Affected users: [NUMBER] ([PERCENTAGE]% of total)
- Service availability: [PERCENTAGE]%
- Estimated resolution: [TIME/UNKNOWN]

WHAT'S HAPPENING NOW:
[Current activities in plain language]

NEXT UPDATE:
[TIME] UTC (in X hours) or when significant progress is made.

Continue tracking: https://status.aura-network.io/incidents/[INC-XXX]

Aura Operations Team
```

### Incident Resolution Notification

```
Subject: [RESOLVED] Aura Network Incident [INC-XXX]

Dear Aura Community,

We are pleased to report that the security incident affecting Aura Network has been resolved.

RESOLUTION SUMMARY:
- Incident detected: [TIME] UTC
- Resolution completed: [TIME] UTC
- Total duration: [DURATION]
- Root cause: [BRIEF EXPLANATION]

WHAT WAS AFFECTED:
- Systems: [LIST]
- Users impacted: [NUMBER]
- Downtime: [DURATION]
- Financial impact: None / [DETAILS]

HOW WE FIXED IT:
- [Action 1]
- [Action 2]
- [Action 3]

USER ACTIONS REQUIRED:
[NONE / List any required actions]

PREVENTING FUTURE INCIDENTS:
We have implemented the following measures:
- [Prevention 1]
- [Prevention 2]
- [Prevention 3]

POST-MORTEM:
A detailed technical post-mortem will be published within 7 days at:
https://blog.aura-network.io/postmortem/[INC-XXX]

TIMELINE:
[Detailed timeline of events]

We sincerely apologize for any inconvenience this incident caused. Thank you for your patience and continued trust in Aura Network.

Questions? security@aura-network.io

Aura Security Team
[TIMESTAMP]
```

### Validator Coordination Message

```
🔒 VALIDATOR ALERT - [SEVERITY]

Incident ID: [INC-XXX]
Time: [TIMESTAMP] UTC

SITUATION:
[Brief technical description]

IMMEDIATE ACTION REQUIRED:
1. [Action 1]
2. [Action 2]
3. [Action 3]

COORDINATION:
- Emergency call: [CONFERENCE BRIDGE]
- Join now for real-time coordination

CHAIN STATUS:
- Current: [STATUS]
- Target: [STATUS]
- Coordination required: [YES/NO]

TECHNICAL DETAILS:
[Relevant technical information for validators]

Please confirm receipt and status within 15 minutes.

Reply with: CONFIRMED [YOUR_MONIKER]
```

## Escalation Procedures

### Severity-Based Escalation

```yaml
critical_incident:
  immediate_notification:
    - Security team (PagerDuty)
    - CTO (Phone + SMS)
    - CEO (Phone + SMS)
  within_15_min:
    - All validators (SMS + Telegram)
    - Incident response team (Slack)
  within_30_min:
    - Executive team
    - Board chair (if funds at risk)
  within_1_hour:
    - All users (Email + In-app)
    - Key partners (Email + Phone)
  within_2_hours:
    - Public announcement (Twitter, Blog)

high_incident:
  immediate_notification:
    - Security team
    - On-call DevOps
  within_30_min:
    - CTO
    - Engineering leads
  within_2_hours:
    - Affected users
    - Relevant validators
  within_4_hours:
    - Partners
    - Public status update

medium_incident:
  within_1_hour:
    - Security team
    - Relevant service owners
  within_4_hours:
    - CTO
    - Status page update
  within_24_hours:
    - Affected users (if any)

low_incident:
  within_4_hours:
    - Relevant team members
  within_24_hours:
    - Status page note
```

### Communication Approval Chain

```
Critical Incidents:
1. Security Lead drafts message
2. CTO reviews and approves
3. CEO approves (for fund-related incidents)
4. Communications team sends

High Incidents:
1. Incident Commander drafts
2. Security Lead approves
3. Communications team sends

Medium/Low Incidents:
1. Relevant team lead drafts
2. Operations lead approves
3. Send via appropriate channels
```

## Update Cadence

### During Active Incident

```python
def calculate_update_frequency(incident):
    if incident.severity == 'critical':
        return timedelta(minutes=30)
    elif incident.severity == 'high':
        return timedelta(hours=2)
    elif incident.severity == 'medium':
        return timedelta(hours=8)
    else:
        return timedelta(hours=24)

def schedule_updates(incident):
    frequency = calculate_update_frequency(incident)
    next_update = datetime.now() + frequency

    while incident.status not in ['resolved', 'closed']:
        if datetime.now() >= next_update:
            send_status_update(incident)
            next_update += frequency
        sleep(60)
```

### Minimum Update Content

Each update must include:
1. Current status
2. Progress since last update
3. Current actions
4. Impact assessment
5. Next expected update time

## FAQ Management

### Pre-Prepared Q&A

#### Q: Are my funds safe?
**A**: [If yes] Yes, user funds are secure and not affected by this incident. [If investigating] We are actively investigating the impact on user funds and will provide an update within [timeframe].

#### Q: When will services be restored?
**A**: [If known] We expect full restoration by [TIME] UTC. [If unknown] We are working as quickly as possible and will provide an estimated timeline in our next update at [TIME] UTC.

#### Q: What caused this incident?
**A**: [If known] The incident was caused by [BRIEF EXPLANATION]. [If investigating] We are still investigating the root cause and will share findings in our post-mortem report.

#### Q: How will you prevent this in the future?
**A**: We are implementing [SPECIFIC MEASURES] to prevent similar incidents. Full details will be included in our post-mortem report.

#### Q: Do I need to take any action?
**A**: [If yes] Yes, please [SPECIFIC ACTIONS]. [If no] No action is required from users at this time.

#### Q: How can I get updates?
**A**: Follow https://status.aura-network.io for real-time updates. We will also send email notifications to all users.

### Customer Support Briefing

```markdown
# Support Team Briefing - [INC-XXX]

## Incident Summary
[2-3 sentence summary in plain language]

## User Impact
- Services affected: [LIST]
- Estimated affected users: [NUMBER]
- User actions required: [YES/NO - details]

## Key Messages
1. [Key message 1]
2. [Key message 2]
3. [Key message 3]

## Response Scripts
See attached document for approved responses

## Escalation
Escalate immediately if user reports:
- [Escalation trigger 1]
- [Escalation trigger 2]

## Resources
- Status page: https://status.aura-network.io
- Internal updates: #incident-response
- Escalation: @incident-commander
```

## Media Relations

### Media Contact Protocol

1. **All media inquiries** → pr@aura-network.io
2. **No individual responses** without approval
3. **Official statements** only from designated spokespersons
4. **Response timeframe**: Within 4 hours for critical incidents

### Designated Spokespersons
- CEO: Strategic and business impact
- CTO: Technical details and architecture
- CSO: Security measures and response

### Press Release Template

```
FOR IMMEDIATE RELEASE

Aura Network Responds to Security Incident

[CITY, DATE] — Aura Network today announced that it has identified and
resolved a security incident affecting [SYSTEMS]. The incident was detected
on [DATE] at [TIME] UTC and resolved on [DATE] at [TIME] UTC.

[Quote from CEO about commitment to security]

Impact:
- [Impact details]
- [User impact]
- [Financial impact]

Response:
- [Response actions]
- [Prevention measures]

[Quote from CTO about technical response]

About Aura Network:
[Boilerplate]

Contact:
[Media contact information]

###
```

## Metrics and Monitoring

### Communication Metrics

```yaml
metrics:
  response_time:
    critical: "< 15 minutes to first notification"
    high: "< 1 hour to first notification"
    target: "95% compliance"

  update_frequency:
    critical: "Every 30 minutes"
    high: "Every 2 hours"
    target: "100% compliance"

  channel_coverage:
    email_delivery: "> 99%"
    status_page_uptime: "> 99.9%"
    social_media_reach: "Track engagement"

  stakeholder_satisfaction:
    user_survey: "Post-incident survey"
    target: "> 70% satisfied with communication"
```

### Communication Dashboard

```bash
# Real-time communication status
┌─────────────────────────────────────────┐
│   Incident Communication Dashboard      │
├─────────────────────────────────────────┤
│ Incident: INC-123                       │
│ Status: Investigating                   │
│ Duration: 2h 15m                        │
│                                         │
│ NOTIFICATIONS SENT:                     │
│ ✅ Internal team    (00:00:05)          │
│ ✅ Validators       (00:00:12)          │
│ ✅ Status page      (00:00:30)          │
│ ✅ Email users      (00:45:00)          │
│ ✅ Social media     (01:00:00)          │
│                                         │
│ NEXT UPDATE DUE: 15 minutes             │
│                                         │
│ COVERAGE:                               │
│ Email delivered: 98.5%                  │
│ Status page views: 15,234               │
│ Twitter impressions: 45,678             │
└─────────────────────────────────────────┘
```

## Post-Incident Communication Review

### Review Checklist

- [ ] Timeline of all communications accurate
- [ ] All stakeholder groups reached appropriately
- [ ] Update frequency met requirements
- [ ] Messages consistent across channels
- [ ] FAQ adequately addressed user questions
- [ ] Media inquiries handled properly
- [ ] Lessons learned documented
- [ ] Templates updated based on feedback

### Communication Retrospective Questions

1. Was our first notification timely enough?
2. Did we provide sufficient detail in updates?
3. Were any stakeholder groups missed?
4. Did our messaging remain consistent?
5. What communication gaps existed?
6. How can we improve for next time?

---

**Document Version**: 1.0
**Last Updated**: 2025-01-13
**Owner**: Communications Lead
**Review Schedule**: Quarterly
