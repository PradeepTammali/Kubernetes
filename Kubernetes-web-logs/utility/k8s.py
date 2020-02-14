# -*- coding=utf-8 -*-
# Copyright 2018 Alex Ma

"""
:author Alex Ma
:date 2018/10/15 

"""
import os
import threading
from utility.log import log
from kubernetes import client, config, watch
from kubernetes.client.rest import ApiException
#from kubernetes.stream import stream

# from kubernetes.client import *
# from kubernetes.client.rest import ApiException


class KubernetesAPI(object):

    def __init__(self, api_host, ssl_ca_cert, key_file, cert_file):
        kub_conf = client.Configuration()
        kub_conf.host = api_host
        kub_conf.ssl_ca_cert = ssl_ca_cert
        kub_conf.cert_file = cert_file
        kub_conf.key_file = key_file
        config.load_kube_config()

        # self.api_client = client.ApiClient(configuration=kub_conf)
        #self.client_core_v1 = client.CoreV1Api(api_client=self.api_client)
        #self.client_apps_v1 = client.AppsV1Api(api_client=self.api_client)
        #self.client_extensions_v1 = client.ExtensionsV1beta1Api(
        #    api_client=self.api_client)

        self.client_core_v1 = client.CoreV1Api()
        self.client_apps_v1 = client.AppsV1Api()
        self.client_extensions_v1 = client.ExtensionsV1beta1Api()

        self.api_dict = {}

    def __getattr__(self, item):
        if item in self.api_dict:
            return self.api_dict[item]
        if hasattr(client, item) and callable(getattr(client, item)):
            self.api_dict[item] = getattr(client, item)(
                api_client=self.api_client)
            return self.api_dict[item]


class K8SClient(KubernetesAPI, threading.Thread):

    def __init__(self, api_host, ssl_ca_cert, key_file, cert_file, ws, namespace, pod, container):
        KubernetesAPI.__init__(self,api_host, ssl_ca_cert, key_file, cert_file)
        threading.Thread.__init__(self)
        self.ws = ws
        self.namespace = namespace
        self.pod = pod
        self.container = container

    @staticmethod
    def gen_ca():
        return None, None, None
        ssl_ca_cert = os.path.join(
            os.path.dirname(os.path.dirname(__file__)),
            '_credentials/kubernetes_dev_ca_cert')
        key_file = os.path.join(
            os.path.dirname(os.path.dirname(__file__)),
            '_credentials/kubernetes_dev_key')
        cert_file = os.path.join(
            os.path.dirname(os.path.dirname(__file__)),
            '_credentials/kubernetes_dev_cert')

        return ssl_ca_cert, key_file, cert_file

    def write_logs(self):
        try:
            import pdb; pdb.set_trace()
            container_logs = self.client_core_v1.read_namespaced_pod_log(name=self.pod,namespace=self.namespace,container=self.container,pretty=True,follow=False,previous=False,timestamps=True,tail_lines=2)
            try:
                self.ws.send(container_logs)
                self.ws.close()
            except Exception as err:
                log.error('container stream err: {}'.format(err))
                self.ws.close()
                return False
        except ApiException as apierr:
            log.error('API Exception while connectin to pod: {}'.format(apierr))
            self.ws.send(apierr.body)
            self.ws.close()
            return False
        except Exception as e:
            log.error('Error while connecting to pod: {}'.format(e))
            self.ws.send("Error while connecting to pod.")
            self.ws.close()
            return False
        finally:
            self.ws.close()
        return True

    def run(self):
        w = watch.Watch()
        try:
            for e in w.stream(self.client_core_v1.read_namespaced_pod_log,name=self.pod,namespace=self.namespace,container=self.container,follow=False,previous=False,timestamps=True):
                if not self.ws.closed:
                    try:
                        self.ws.send(e)
                    except Exception as err:
                        log.error('container stream err: {}'.format(err))
                        w.stop()
                        self.ws.close()
                        break
        except ApiException as apierr:
            log.error('API Exception while connectin to pod: {}'.format(apierr))
            self.ws.send(apierr.body)
            w.stop()
            self.ws.close()
        except Exception as e:
            log.error('Error while connecting to pod: {}'.format(e))
            self.ws.send("Error while connecting to pod.")
            w.stop()
            self.ws.close()
        finally:
            w.stop()
            self.ws.close()


'''
class K8SStreamThread(threading.Thread):

    def __init__(self, ws, container_stream):
        super(K8SStreamThread, self).__init__()
        self.ws = ws
        self.stream = container_stream

    def run(self):
        while not self.ws.closed:
            import pdb; pdb.set_trace()
            if self.stream.isclosed():
                log.info('container stream closed')
                self.ws.close()
            for e in self.stream.stream():
                try:
                    self.ws.send(e)
                    if self.stream.peek_stdout():
                        stdout = self.stream.read_stdout()
                        self.ws.send(stdout)

                    if self.stream.peek_stderr():
                        stderr = self.stream.read_stderr()
                        self.ws.send(stderr)
                except Exception as err:
                    log.error('container stream err: {}'.format(err))
                    self.ws.close()
                    break
'''