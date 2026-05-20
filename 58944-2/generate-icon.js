const fs = require('fs');

function createPNG(width, height, r, g, b, a) {
  const signature = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);

  function crc32(buf) {
    let crc = -1;
    for (let i = 0; i < buf.length; i++) {
      crc ^= buf[i];
      for (let j = 0; j < 8; j++) {
        crc = (crc >>> 1) ^ (crc & 1 ? 0xedb88320 : 0);
      }
    }
    return (crc ^ -1) >>> 0;
  }

  function createChunk(type, data) {
    const length = Buffer.alloc(4);
    length.writeUInt32BE(data.length, 0);
    const typeBuf = Buffer.from(type);
    const crcBuf = Buffer.alloc(4);
    crcBuf.writeUInt32BE(crc32(Buffer.concat([typeBuf, data])), 0);
    return Buffer.concat([length, typeBuf, data, crcBuf]);
  }

  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(width, 0);
  ihdr.writeUInt32BE(height, 4);
  ihdr[8] = 8;
  ihdr[9] = 6;
  ihdr[10] = 0;
  ihdr[11] = 0;
  ihdr[12] = 0;

  const rawData = [];
  for (let y = 0; y < height; y++) {
    rawData.push(0);
    for (let x = 0; x < width; x++) {
      const cx = width / 2;
      const cy = height / 2;
      const dist = Math.sqrt((x - cx) ** 2 + (y - cy) ** 2);
      const maxDist = Math.min(width, height) / 2 - 2;
      
      if (dist < maxDist) {
        const gradient = 1 - (dist / maxDist) * 0.3;
        rawData.push(Math.round(r * gradient));
        rawData.push(Math.round(g * gradient));
        rawData.push(Math.round(b * gradient));
        rawData.push(a);
      } else {
        rawData.push(0);
        rawData.push(0);
        rawData.push(0);
        rawData.push(0);
      }
    }
  }

  const zlib = require('zlib');
  const compressed = zlib.deflateSync(Buffer.from(rawData));
  const idat = compressed;

  return Buffer.concat([
    signature,
    createChunk('IHDR', ihdr),
    createChunk('IDAT', idat),
    createChunk('IEND', Buffer.alloc(0))
  ]);
}

const sizes = [16, 48, 128];
const colors = [
  { r: 102, g: 126, b: 234 },
  { r: 118, g: 75, b: 162 }
];

sizes.forEach(size => {
  const png = createPNG(size, size, colors[0].r, colors[0].g, colors[0].b, 255);
  fs.writeFileSync(`icon${size === 16 ? '' : size}.png`, png);
  console.log(`Created ${size}x${size} icon`);
});

console.log('Icons generated successfully!');
