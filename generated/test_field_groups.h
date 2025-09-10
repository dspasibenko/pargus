
#include <Arduino.h>
#include "bigendian.h"

namespace test {


// Read-write register with mixed field types
static constexpr int Reg_RW_ID = 0;
struct RW {

    // Send read-only fields to wire (for reading data from device)
    int send_read_data(uint8_t* buf, size_t size) {
        int written = 0;
        written += bigendian::encode(buf + written, this->rw_field1);
        written += bigendian::encode(buf + written, this->rw_field2);
        written += bigendian::encode(buf + written, this->read_field1);
        written += bigendian::encode(buf + written, this->read_field2);
        written += bigendian::encode(buf + written, this->write_field1);
        written += bigendian::encode(buf + written, this->write_field2);
        return written;
    }

    // Send write-only fields to wire (for writing data to device)
    int send_write_data(uint8_t* buf, size_t size) {
        int written = 0;
        written += bigendian::encode(buf + written, this->rw_field1);
        written += bigendian::encode(buf + written, this->rw_field2);
        written += bigendian::encode(buf + written, this->read_field1);
        written += bigendian::encode(buf + written, this->read_field2);
        written += bigendian::encode(buf + written, this->write_field1);
        written += bigendian::encode(buf + written, this->write_field2);
        return written;
    }

    // Get read-only fields from wire (for updating data from device)
    int receive_read_data(uint8_t* buf, size_t size) {
        int read = 0;
        read += bigendian::decode(this->rw_field1, buf + read);
        read += bigendian::decode(this->rw_field2, buf + read);
        read += bigendian::decode(this->read_field1, buf + read);
        read += bigendian::decode(this->read_field2, buf + read);
        read += bigendian::decode(this->write_field1, buf + read);
        read += bigendian::decode(this->write_field2, buf + read);
        return read;
    }

    // Getting write-only fields from wire (for getting write commands)
    int receive_write_data(uint8_t* buf, size_t size) {
        int read = 0;
        read += bigendian::decode(this->rw_field1, buf + read);
        read += bigendian::decode(this->rw_field2, buf + read);
        read += bigendian::decode(this->read_field1, buf + read);
        read += bigendian::decode(this->read_field2, buf + read);
        read += bigendian::decode(this->write_field1, buf + read);
        read += bigendian::decode(this->write_field2, buf + read);
        return read;
    }
};

// Read-only register
static constexpr int Reg_R_ID = 1;
struct R {
    // Read-only fields
    // Read-only field
    uint8_t status;
    // Read-only field
    int32_t counter;
    // Read-only field
    // Bit field: flags
    static constexpr uint8_t bit0_bm = 0x1; // bit 0
    static constexpr uint8_t bit15_bm = 0x3E; // bits 1-5
    uint8_t flags;

    // Send read-only fields to wire (for reading data from device)
    int send_read_data(uint8_t* buf, size_t size) {
        int written = 0;
        written += bigendian::encode(buf + written, this->status);
        written += bigendian::encode(buf + written, this->counter);
        written += bigendian::encode(buf + written, this->flags);
        return written;
    }

    // Send write-only fields to wire (for writing data to device)
    int send_write_data(uint8_t* buf, size_t size) {
        return -1; // read-only register has no write data
    }

    // Get read-only fields from wire (for updating data from device)
    int receive_read_data(uint8_t* buf, size_t size) {
        int read = 0;
        read += bigendian::decode(this->status, buf + read);
        read += bigendian::decode(this->counter, buf + read);
        read += bigendian::decode(this->flags, buf + read);
        return read;
    }

    // Getting write-only fields from wire (for getting write commands)
    int receive_write_data(uint8_t* buf, size_t size) {
        return -1; // read-only register cannot receive write data
    }
};

// Write-only register
static constexpr int Reg_W_ID = 2;
struct W {
    // Write-only fields
    // Write-only field
    uint16_t command;
    // Write-only field
    int8_t value;
    // Write-only field
    // Bit field: config
    static constexpr uint8_t bit0_bm = 0x1; // bit 0
    static constexpr uint8_t bit23_bm = 0xC; // bits 2-3
    uint8_t config;

    // Send read-only fields to wire (for reading data from device)
    int send_read_data(uint8_t* buf, size_t size) {
        return -1; // write-only register has no read data
    }

    // Send write-only fields to wire (for writing data to device)
    int send_write_data(uint8_t* buf, size_t size) {
        int written = 0;
        written += bigendian::encode(buf + written, this->command);
        written += bigendian::encode(buf + written, this->value);
        written += bigendian::encode(buf + written, this->config);
        return written;
    }

    // Get read-only fields from wire (for updating data from device)
    int receive_read_data(uint8_t* buf, size_t size) {
        return -1; // write-only register cannot receive read data
    }

    // Getting write-only fields from wire (for getting write commands)
    int receive_write_data(uint8_t* buf, size_t size) {
        int read = 0;
        read += bigendian::decode(this->command, buf + read);
        read += bigendian::decode(this->value, buf + read);
        read += bigendian::decode(this->config, buf + read);
        return read;
    }
};


} // namespace test
