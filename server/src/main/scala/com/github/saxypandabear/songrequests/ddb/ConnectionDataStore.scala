package com.github.saxypandabear.songrequests.ddb

import com.github.saxypandabear.songrequests.ddb.model.Connection

/**
 * A data store that manages details about a user's connection to our services
 * This is basically an abstraction around our database usage
 */
trait ConnectionDataStore {
    /**
     * Fetch the connection details for a user by their channel ID, which is
     * the primary key (or hash key for DynamoDB).
     * This should always get the most up-to-date value of the data (a consistent read)
     * @param channelId the Twitch channel ID
     * @return a POJO that represents the connection document
     * @throws RuntimeException when the channelId does not exist in the data store
     */
    def getConnectionDetailsById(channelId: String): Connection

    /**
     * Update a record in the data store with the given hash key, and the given
     * object.
     * @param channelId  the Twitch channel ID
     * @param connection connection object
     * @throws RuntimeException when the channelId does not exist in the data store,
     *                          or the connection object is malformed
     */
    def updateConnectionDetailsById(channelId: String, connection: Connection): Unit
}
