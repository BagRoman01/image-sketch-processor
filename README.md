# image-sketch-processor
Image Sketch Processor - Асинхронный сервис преобразования изображений в рисунки
docker run -d --name rustfs -p 9000:9000 -e RUSTFS_ADDRESS=0.0.0.0:9000
 -e RUSTFS_ACCESS_KEY=admin -e RUSTFS_SECRET_KEY=admin 
 -v D:/GoProjects/data:/data -v D:/GoProjects/logs:/app/logs rustfs/rustfs:alpha