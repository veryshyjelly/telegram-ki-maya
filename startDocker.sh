sudo docker build -t telegram-ki-maya .
sudo docker stop telegram-ki-maya && sudo docker rm telegram-ki-maya
sudo docker run -d -p 8060:8060 --name telegram-ki-maya telegram-ki-maya