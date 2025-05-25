package interfaces

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
	"github.com/gin-gonic/gin"
)

func GenClusterIP() string {
	cip := "222.111." + strconv.Itoa(
		rand.Intn(256),
	) + "." + strconv.Itoa(
		rand.Intn(256),
	)

	return cip
}

func CreateService(c *gin.Context) {
	// 	var svc object.Service
	// 	if err := c.BindJSON(&svc); err != nil {
	// 		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
	// 		return
	// 	}

	// 	if svc.Metadata.Namespace == "" {
	// 		svc.Metadata.Namespace = DefaultNamespace
	// 	}

	// 	st, _ := object.NewServiceStore([]string{})

	// 	defer func() {
	// 		if closeErr := st.Close(); closeErr != nil {
	// 			c.JSON(
	// 				http.StatusInternalServerError,
	// 				"Failed to close service store: "+closeErr.Error(),
	// 			)
	// 		}
	// 	}()

	// 	// Check if the service already exists
	// 	reply, err := st.GetService(
	// 		c.Request.Context(),
	// 		svc.Metadata.Namespace,
	// 		svc.Metadata.Name,
	// 	)
	// 	if err != nil {
	// 		c.JSON(
	// 			http.StatusInternalServerError,
	// 			"Failed to get service: "+err.Error(),
	// 		)

	// 		return
	// 	}

	// 	if reply != nil {
	// 		c.JSON(
	// 			http.StatusConflict,
	// 			"Service already exists: "+svc.Metadata.Name,
	// 		)

	// 		return
	// 	}

	// 	// Create a new ClusterIP
	// 	var clusterIPStore *object.ClusterIPStore
	// 	clusterIPStore, _ = object.NewClusterIPStore([]string{})

	// 	defer func() {
	// 		if closeErr := clusterIPStore.Close(); closeErr != nil {
	// 			c.JSON(
	// 				http.StatusInternalServerError,
	// 				"Failed to close cluster IP store: "+closeErr.Error(),
	// 			)
	// 		}
	// 	}()

	// 	var clusterIP string
	// 	// 反复生成ClusterIP，直到生成的ClusterIP不在etcd中
	// 	for {
	// 		clusterIP = GenClusterIP()
	// 		res, _ := clusterIPStore.GetClusterIP(
	// 			c.Request.Context(),
	// 			clusterIP,
	// 		)

	// 		if res != clusterIP {
	// 			// 存储ClusterIP
	// 			_ = clusterIPStore.SetClusterIP(
	// 				c.Request.Context(),
	// 				clusterIP,
	// 			)
	// 			// 存储成功，跳出循环
	// 			break
	// 		}
	// 	}

	// 	svc.Status.ClusterIP = clusterIP

	// 	// Apply the service to etcd
	// 	_ = st.AddService(c.Request.Context(), &svc)

	// 	// 发送给消息队列
	// 	jsonData, _ := json.Marshal(svc)
	// 	_ = mqtemplate.SendMessageToQueue(
	// 		mqtemplate.CreateServiceQueueName,
	// 		string(jsonData),
	// 	)

	// 	path := path.Join(etcd.ClusterIPPrefix, clusterIP)
	// 	c.JSON(http.StatusOK, "Service created: "+path)
}

func DeleteService(c *gin.Context) {
	// 	var svc object.Service
	// 	if err := c.BindJSON(&svc); err != nil {
	// 		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
	// 		return
	// 	}

	// 	if svc.Metadata.Namespace == "" {
	// 		svc.Metadata.Namespace = DefaultNamespace
	// 	}

	// 	st, _ := object.NewServiceStore([]string{})

	// 	defer func() {
	// 		if closeErr := st.Close(); closeErr != nil {
	// 			c.JSON(
	// 				http.StatusInternalServerError,
	// 				"Failed to close service store: "+closeErr.Error(),
	// 			)
	// 		}
	// 	}()

	// 	reply, err := st.GetService(
	// 		c.Request.Context(),
	// 		svc.Metadata.Namespace,
	// 		svc.Metadata.Name,
	// 	)
	// 	if err != nil {
	// 		c.JSON(
	// 			http.StatusInternalServerError,
	// 			"Failed to get service: "+err.Error(),
	// 		)

	// 		return
	// 	}

	// 	if reply == nil {
	// 		c.JSON(
	// 			http.StatusNotFound,
	// 			"Service not found: "+svc.Metadata.Name,
	// 		)

	// 		return
	// 	}

	// 	err = st.DeleteService(
	// 		c.Request.Context(),
	// 		svc.Metadata.Namespace,
	// 		svc.Metadata.Name,
	// 	)
	// 	if err != nil {
	// 		c.JSON(
	// 			http.StatusInternalServerError,
	// 			"Failed to delete service: "+err.Error(),
	// 		)

	// 		return
	// }

	// c.JSON(http.StatusOK, "Service deleted: "+svc.Metadata.Name)
}

func GetService(c *gin.Context) {
	// 	var svc object.Service
	// 	if err := c.BindJSON(&svc); err != nil {
	// 		c.JSON(http.StatusBadRequest, "Invalid JSON: "+err.Error())
	// 		return
	// 	}

	// 	if svc.Metadata.Namespace == "" {
	// 		svc.Metadata.Namespace = "default"
	// 	}

	// 	st, _ := object.NewServiceStore([]string{})

	// 	defer func() {
	// 		if closeErr := st.Close(); closeErr != nil {
	// 			c.JSON(
	// 				http.StatusInternalServerError,
	// 				"Failed to close service store: "+closeErr.Error(),
	// 			)
	// 		}
	// 	}()

	// 	pods, err := st.GetService(
	// 		c.Request.Context(),
	// 		svc.Metadata.Namespace,
	// 		svc.Metadata.Name,
	// 	)
	// 	if err != nil {
	// 		c.JSON(
	// 			http.StatusInternalServerError,
	// 			"Failed to get service: "+err.Error(),
	// 		)

	// 		return
	// 	}

	// 	if pods == nil {
	// 		c.JSON(
	// 			http.StatusNotFound,
	// 			"Service not found: "+svc.Metadata.Name,
	// 		)

	//		return
	//	}
}

func GetAllService(c *gin.Context) {
	st, _ := object.NewServiceStore([]string{})

	defer func() {
		if closeErr := st.Close(); closeErr != nil {
			c.JSON(
				http.StatusInternalServerError,
				"Failed to close service store: "+closeErr.Error(),
			)
		}
	}()

	svcs, err := st.ListServices(c.Request.Context())
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			"Failed to get all services: "+err.Error(),
		)

		return
	}

	c.JSON(http.StatusOK, svcs)
}
