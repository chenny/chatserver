package gotest;

import com.chatserver.client.AesHelper;
import com.google.protobuf.InvalidProtocolBufferException;
import pb.*;

public class GoTest {
	public static void main(String[] args) {
		// 定义client，并连接
		com.chatserver.client.TcpHelper client = new com.chatserver.client.TcpHelper();
		boolean re = client.connect("127.0.0.1", 8989);
		if (!re) {
			System.out.println("connect error\n");
			return;
		}
		
		// 初始化登录包
		Pb.PbClientLogin.Builder clientLoginBuilder = Pb.PbClientLogin.newBuilder();
		clientLoginBuilder.setUuid("hello， 我是一个GG, 正在写一个java程序");
		clientLoginBuilder.setVersion(3.14f);
		clientLoginBuilder.setTimestamp((int)(System.currentTimeMillis()/1000));
		
		// 序列化登录包，并发送
		byte[] sendData = marshalPbClientLogin(clientLoginBuilder);
		re = client.send(sendData);
		if (!re) {
			System.out.println("send error\n");
			return;
		}
		
		// 读取数据
		byte[] readData = new byte[1024];
		int readSize = client.read(readData);
		if (-1 == readSize) {
			System.out.println("read error\n");
			return;
		}
		
		// 假定读取的数据是登录包类型
		byte[] tempData = new byte[readSize];
		System.arraycopy(readData, 0, tempData, 0, readSize);
		Pb.PbClientLogin clientLogin = unmarshalPbClientLogin(tempData);
		if (null == clientLogin) {
			System.out.println("unmarshalPbClientLogin error\n");
			return;
		}
		System.out.println(clientLogin.getUuid());
		System.out.println(clientLogin.getVersion());
		System.out.println(clientLogin.getTimestamp());
		
		// 关闭连接
		client.close();
	}
	
	// 将pb结构序列化，再aes加密
	public static byte[] marshalPbClientLogin(Pb.PbClientLogin.Builder clientLoginBuilder) {
		AesHelper aes = new AesHelper("12345abcdef67890");
		Pb.PbClientLogin info = clientLoginBuilder.build();
		return aes.encrypt(info.toByteArray());
	}
	
	// 先aes解密，再反序列化为pb结构
	public static Pb.PbClientLogin unmarshalPbClientLogin(byte[] data) {
		AesHelper aes = new AesHelper("12345abcdef67890");
		byte[] decrypted = aes.decrypt(data);
		
		Pb.PbClientLogin rev = null;
		try {
			rev = Pb.PbClientLogin.parseFrom(decrypted);
		} catch (InvalidProtocolBufferException e) {
			return null;
		}
		return rev;
	}
}
