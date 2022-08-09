package k8s

import (
	"flag"
	"github.com/gorilla/websocket"
	"k8s.io/client-go/kubernetes"
	resetclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"path/filepath"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type KubeLogger struct {
	Conn *websocket.Conn
}

func SbxInitClientSet() *kubernetes.Clientset {
	var sbxkubeConfig *string
	if home := homeDir(); home != "" {
		sbxkubeConfig = flag.String("sbxkubeConfig", filepath.Join(home, ".kube", "config1"), "(optional) absolute path to the kubeConfig file")
	} else {
		sbxkubeConfig = flag.String("sbxkubeConfig", "", "absolute path to the kubeConfig file")
	}
	flag.Parse()

	sbxconfig, err := resetclient.InClusterConfig()
	if err != nil {
		sbxconfig, err = clientcmd.BuildConfigFromFlags("", *sbxkubeConfig)
		if err != nil {
			panic(err)
		}
	}
	sbxclientSet, err := kubernetes.NewForConfig(sbxconfig)
	if err != nil {
		panic(err)
	}

	return sbxclientSet
}

func SitInitClientSet() *kubernetes.Clientset {
	var sitkubeConfig *string
	if home := homeDir(); home != "" {
		sitkubeConfig = flag.String("sitkubeConfig", filepath.Join(home, ".kube", "config2"), "(optional) absolute path to the kubeConfig file")
	} else {
		sitkubeConfig = flag.String("sitkubeConfig", "", "absolute path to the kubeConfig file")
	}
	flag.Parse()

	sitconfig, err := resetclient.InClusterConfig()
	if err != nil {
		sitconfig, err = clientcmd.BuildConfigFromFlags("", *sitkubeConfig)
		if err != nil {
			panic(err)
		}
	}
	sitclientSet, err := kubernetes.NewForConfig(sitconfig)
	if err != nil {
		panic(err)
	}

	return sitclientSet
}

func homeDir() string {
	home := os.Getenv("HOME")
	return home
}

func NewKubeLogger(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*KubeLogger, error) {
	conn, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	KubeLogger := &KubeLogger{
		Conn: conn,
	}
	return KubeLogger, nil
}

func (kl *KubeLogger) Write(data []byte) error {
	if err := kl.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return  err
	}
	return nil
}

func (kl *KubeLogger) Close() error {
	return kl.Conn.Close()
}