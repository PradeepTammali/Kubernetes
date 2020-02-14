
from werkzeug import serving
from flask_sockets import Sockets
from flask import Flask, render_template
from utility.log import log
from utility.k8s import K8SClient

app = Flask(__name__, static_folder='static',
            static_url_path='/terminal/static')
sockets = Sockets(app)


@app.route('/logs/', methods=['GET'])
def index():
    return render_template('index.html')


@app.route('/logs/window', methods=['GET'])
def terminal():
    return render_template('terminal.html')


@sockets.route('/logs/<namespace>/<pod>/<container>')
def terminal_socket(ws, namespace, pod, container):
    log.info('Try create socket connection')
    ssl_ca_cert, key_file, cert_file = K8SClient.gen_ca()
    kub = K8SClient(
        api_host='https://10.0.75.1:6443',
        ssl_ca_cert=ssl_ca_cert,
        key_file=key_file,
        cert_file=cert_file,
        ws=ws,
        namespace=namespace,
        pod=pod,
        container=container)

    log.info('Start logs')
    kub.write_logs()
    #kub.start()
    try:
        while not ws.closed:
            message = ws.receive()
            if message is not None:
                if message != '__ping__':
                    log.info(message)
        log.info('Connection Closed.\r')
    except Exception as err:
        log.error('Connect container error: {}'.format(err))
    finally:
        ws.close()
    

@serving.run_with_reloader
def run_server():
    app.debug = True
    from gevent import pywsgi
    from geventwebsocket.handler import WebSocketHandler
    server = pywsgi.WSGIServer(
        listener = ('0.0.0.0', 5000),
        application=app,
        handler_class=WebSocketHandler)
    server.serve_forever()


if __name__ == '__main__':
    run_server()
