HTTPS_ONLY = True
DOMAIN_PREFIX = 'dev'  # Change it to `local` when running in laptop
# Only have the services you need to run in the set below.
SERVICES = {
    "search",
    "usersearch",
    "collab",
    "finance",
    "financeuser",
    "scraping",
    "graph",
    "usergraph",
    "gateway",
    "user",
    "userprofile",
    "lookup",
    "login",
    "web"
}

### ALL Services. Don't modify this comment, this is just here to list all services
#   whenever you modify the above `SERVICES` set
'''
SERVICES = {
    "search",
    "usersearch",
    "collab",
    "finance",
    "scraping",
    "graph",
    "usergraph",
    "gateway",
    "user",
    "userprofile",
    "lookup",
    "login",
    "web"
}
'''
###


############################################################
import yaml

default = '''
global
    stats socket /var/run/api.sock user root group root mode 660 level admin expose-fd listeners
    log stdout format raw local0 info

defaults
    mode http
    timeout client 100s
    timeout connect 50s
    timeout server 100s
    timeout http-request 100s
    log global

frontend rest
    bind *:80'''

if HTTPS_ONLY:
    default += '''
    bind *:443 ssl crt /etc/haproxy/certs/sentieo.com.pem
    redirect scheme https if !{ ssl_fc }

    log-format "%{+Q}o\ backend = %b, path, = %HU, status = %ST"
    log stdout local0 debug
'''

default += f'''
    http-response set-header Access-Control-Allow-Origin "{'https' if HTTPS_ONLY else 'http'}://{DOMAIN_PREFIX}.sentieo.com"
    http-response set-header Access-Control-Allow-Headers "DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,X-CSRFToken"
    http-response set-header Access-Control-Allow-Methods "GET, POST, OPTIONS, PUT, DELETE, PATCH"
    http-response set-header Access-Control-Allow-Credentials true
    http-response set-header Access-Control-Expose-Headers "Content-Length,Content-Range"
    http-response set-header Access-Control-Max-Age 3628800
'''

append = '''

frontend stats
    bind                0.0.0.0:8888
    mode            	http
    stats           	enable
    option          	httplog
    stats           	show-legends
    stats          		uri /haproxy
    stats           	realm Haproxy\ Statistics
    stats           	refresh 5s
'''

domainMap = {
    'search' : {
        'port': 8080,
        'dns': [
            f'{DOMAIN_PREFIX}-search'
        ]
    },
    'usersearch' : {
        'port': 8081,
        'dns': [
            f'user-{DOMAIN_PREFIX}-search'
        ]
    },
    'collab' : {
        'port': 8200,
        'dns': [
            f'user-{DOMAIN_PREFIX}-rms',
            f'user-{DOMAIN_PREFIX}-settings'
        ]
    },
    'finance' : {
        'port': 8080,
        'dns': [
            f'{DOMAIN_PREFIX}-finance',
        ]
    },
    'financeuser' : {
        'port': 8080,
        'dns': [
            f'user-{DOMAIN_PREFIX}-finance'
        ]
    },
    'scraping' : {
        'port': 8080,
        'dns': [
            f'{DOMAIN_PREFIX}-scraping'
        ]
    },
    'graph' : {
        'port': 8080,
        'dns': [
            f'{DOMAIN_PREFIX}-graph',
        ]
    },
    'usergraph' : {
        'port': 8080,
        'dns': [
            f'user-{DOMAIN_PREFIX}-graph'
        ]
    },
    'gateway' : {
        'port': 8080,
        'dns': [
            f'{DOMAIN_PREFIX}-gateway'
        ]
    },
    'user' : {
        'port': 5000,
        'dns': [
            f'user-{DOMAIN_PREFIX}-mgmt'
        ]
    },
    'userprofile' : {
        'port': 8000,
        'dns': [
            f'{DOMAIN_PREFIX}-userprofile',
        ]
    },
    'lookup' : {
        'port': 8000,
        'dns': [
            f'{DOMAIN_PREFIX}-lookup'
        ]
    },
    'login' : {
        'port': 80,
        'dns': [
            f'user-{DOMAIN_PREFIX}'
        ]
    },
    'web' : {
        'port': 80,
        'dns': [
            f'{DOMAIN_PREFIX}'
        ]
    },
}

composeCfg = {
    "version": "3.8",
    "services": {
        "haproxy": {
            "container_name": "proxy",
            "image": "haproxytech/haproxy-alpine:2.6",
            "command": "haproxy -f /etc/haproxy/",
            "ports": [
                "80:80"
            ],
            "volumes": [
                "./haproxy.cfg:/etc/haproxy/haproxy.cfg"
            ]
        }
    }
}

if HTTPS_ONLY:
    composeCfg['services']['haproxy']['ports'].append("443:443")
    composeCfg['services']['haproxy']['volumes'].append("./sentieo.com.pem:/etc/haproxy/certs/sentieo.com.pem")

haproxyCfg = default
servers = ''
reqConditionsSoFar = ''
domainsList = f'127.0.0.1   {DOMAIN_PREFIX}.sentieo.com'

for file, desc in domainMap.items():
    if file not in SERVICES:
        continue

    composeCfg['services'][f'sentieo{file}'] = {
        'container_name': file,
        'image': f'602037364990.dkr.ecr.ap-south-1.amazonaws.com/sentieo{file if file != "collab" else "rms"}:latest'
    }
    if file == 'collab':
        composeCfg['services'][f'sentieo{file}']['command'] = 'python ../sentieouserwebapp/manage.py runserver 0.0.0.0:8200'
    elif 'graph' in file:
        composeCfg['services'][f'sentieo{file}']['image'] = f'602037364990.dkr.ecr.ap-south-1.amazonaws.com/sentieographs:{"sentieo-latest" if file == "graph" else "user-latest"}'

    if file == "lookup":
        composeCfg['services'][f'sentieo{file}']['environment'] = [
            "COMPANY_ES=local-db.sntio.com",
            "COMPANY_ES_PORT=5200",
            "COMPANY_ES2=local-db.sntio.com",
            "ENV_FOR_DYNACONF=development",
            "GATEWAY_SERVER=http://sentieogateway:8080/",
            "USER_ENTITY_ES=local-db.sntio.com"
        ]

    if 'depends_on' not in composeCfg['services']['haproxy']:
        composeCfg['services']['haproxy']['depends_on'] = []
    composeCfg['services']['haproxy']['depends_on'].append(f'sentieo{file}')

    if file == 'web':
        composeCfg['services'][f'sentieo{file}']['environment'] = [
            "NGINX_ENVSUBST_OUTPUT_DIR=/etc/nginx/sites-available",
            f"NGINX_SERVER={DOMAIN_PREFIX}.sentieo.com"
        ]
        # composeCfg['services'][f'sentieo{file}']['volumes'] = [
        #     './nginx.conf:/etc/nginx/sites-enabled/local.sentieo.com.conf',
        # ]
        composeCfg['services'][f'sentieo{file}']['command'] = "nginx -g 'daemon off;'"
        servers += f'''
backend sentieo{file}-srv
    server {file}-server sentieo{file}:{desc['port']} check
'''
        continue
    with open(f'routes/{file}.yaml', 'r') as f:
        apis = yaml.full_load(f)
        for dns in desc['dns']:
            #LEVEL 2
            haproxyCfg += f'''
    acl {file}_req hdr(host) -i {dns}.sentieo.com'''
            domainsList += f'''
127.0.0.1   {dns}.sentieo.com'''
        if apis and 'apis' in apis:
            for  api in apis['apis']:
                haproxyCfg += f'''
    acl {file}_req path_beg -i /api/{api}'''
        if file != 'login':
            reqConditionsSoFar += f' !{file}_req'
        #LEVEL 1
        servers += f'''
backend sentieo{file}-srv
    server {file}-server sentieo{file}:{desc['port']} check
'''
        #LEVEL 1
        extras = ''
        if file == 'login':
            extras = reqConditionsSoFar
        haproxyCfg += f'''
    use_backend sentieo{file}-srv if {file}_req{extras}
'''
        f.close()

if 'web' in SERVICES:
    haproxyCfg += '''
    default_backend sentieoweb-srv
'''
haproxyCfg += servers
haproxyCfg += append

with open('deployment/haproxy.cfg', 'w') as f:
    f.write(haproxyCfg)
    f.close()

with open('deployment/docker-compose.yml', 'w') as f:
    yaml.dump(composeCfg, f)
    f.close()

print ('Add the following lines to `/etc/hosts` file or equivalent in your laptop')
print (domainsList)
print ("\n\nAfter adding the above domains you can run these commands to get your services up")
print ("$ cd deployment")
print ("$ docker-compose up -d")