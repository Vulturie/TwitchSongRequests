package com.github.saxypandabear.songrequests.metric

import java.util.Date
import java.util.concurrent.ExecutorService

import com.amazonaws.services.cloudwatch.AmazonCloudWatch
import com.amazonaws.services.cloudwatch.model.{
  Dimension,
  MetricDatum,
  PutMetricDataRequest,
  StandardUnit
}
import com.typesafe.scalalogging.LazyLogging

import scala.collection.JavaConverters._

/**
 * Simple implementation that takes data and publishes to CloudWatch.
 * This should perform the operations asynchronously, because we expect
 * to share this object throughout the app, and can't let this metric
 * emitter be the bottleneck for processing data.
 * @param client          internal AWS SDK CloudWatch client
 * @param executorService thread pool executor to submit EmitMetricTasks to
 */
class CloudWatchMetricCollector(
    client: AmazonCloudWatch,
    executorService: ExecutorService
) {

  def emitCountMetric(
      name: String,
      value: Double,
      tags: Map[String, String] = Map.empty
  ): Unit =
    executorService.submit(new EmitMetricTask(client, name, value, tags))
}

class EmitMetricTask(
    client: AmazonCloudWatch,
    name: String,
    value: Double,
    tags: Map[String, String]
) extends Runnable
    with LazyLogging {
  override def run(): Unit = {
    val datum = new MetricDatum()
      .withMetricName(name)
      .withTimestamp(new Date())
      .withUnit(StandardUnit.Count)
      .withValue(value)

    if (tags.nonEmpty) {
      datum.setDimensions(convertMapToDimensions(tags).asJava)
    }

    logger.info("Emitting new metric[{}={}]", name, value)
    val request = new PutMetricDataRequest()
      .withNamespace("TwitchSongRequests")
      .withMetricData(datum)

    val response = client.putMetricData(request)
    logger.info(
        "Response from emitting metric[{}={}]: {}",
        name,
        value,
        response
    )
  }

  private def convertMapToDimensions(
      tags: Map[String, String]
  ): Seq[Dimension] =
    tags.map { case (k, v) =>
      new Dimension().withName(k).withValue(v)
    }.toSeq
}

object MetricsConstants {
  val SONG_REQUEST_RECEIVED = "received-song-request-count"
  val SENT_TO_SQS_TOTAL     = "total-send-sqs-count"
  val SENT_TO_SQS_FAILED    = "failed-send-sqs-count"
  val OAUTH_TOKEN_REFRESHED = "token-refresh-count"
}