FROM python:3.11-slim
WORKDIR /app
COPY trainer.py /app/trainer.py
RUN pip install --no-cache-dir boto3
CMD ["python", "trainer.py"]